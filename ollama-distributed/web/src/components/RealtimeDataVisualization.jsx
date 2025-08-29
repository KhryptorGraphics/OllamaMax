import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { Card, Row, Col, Button, Form, Badge, Modal, Dropdown, Alert } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faChartLine,
  faChartBar,
  faChartPie,
  faChartArea,
  faTh,
  faExpand,
  faCompress,
  faDownload,
  faSync,
  faPlay,
  faPause,
  faStop,
  faCog,
  faFilter,
  faEye,
  faEyeSlash,
  faMaximize,
  faMinimize,
  faPalette,
  faRedo,
  faCamera,
  faShareAlt
} from '@fortawesome/free-solid-svg-icons';
import '../styles/design-system.css';

const RealtimeDataVisualization = ({
  title = 'Real-time Data Visualization',
  data = [],
  metrics = [],
  onDataUpdate,
  onExport,
  updateInterval = 5000,
  maxDataPoints = 100,
  className = '',
  height = 400,
  enableRealtime = true,
  enableExport = true,
  enableFullscreen = true
}) => {
  // Chart state
  const [chartType, setChartType] = useState('line');
  const [isRealtime, setIsRealtime] = useState(enableRealtime);
  const [isPaused, setIsPaused] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [selectedMetrics, setSelectedMetrics] = useState(metrics.map(m => m.id));
  const [timeRange, setTimeRange] = useState('5m');
  const [colorScheme, setColorScheme] = useState('default');
  const [showLegend, setShowLegend] = useState(true);
  const [smoothing, setSmoothing] = useState(false);
  
  // Data management
  const [chartData, setChartData] = useState([]);
  const [aggregatedData, setAggregatedData] = useState([]);
  const [stats, setStats] = useState({});
  
  // UI state
  const [showSettings, setShowSettings] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [alertThresholds, setAlertThresholds] = useState({});
  const [activeAlerts, setActiveAlerts] = useState([]);
  
  // Refs
  const chartContainerRef = useRef(null);
  const canvasRef = useRef(null);
  const intervalRef = useRef(null);
  const dataBufferRef = useRef([]);

  // Color schemes
  const colorSchemes = {
    default: ['#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe'],
    vibrant: ['#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffd93d', '#6c5ce7'],
    professional: ['#2d3748', '#4a5568', '#718096', '#a0aec0', '#cbd5e0', '#e2e8f0'],
    sunset: ['#fd79a8', '#fdcb6e', '#e17055', '#6c5ce7', '#74b9ff', '#55a3ff'],
    nature: ['#00b894', '#00cec9', '#0984e3', '#6c5ce7', '#a29bfe', '#fd79a8']
  };

  // Time range options
  const timeRangeOptions = [
    { value: '1m', label: '1 Minute', seconds: 60 },
    { value: '5m', label: '5 Minutes', seconds: 300 },
    { value: '15m', label: '15 Minutes', seconds: 900 },
    { value: '1h', label: '1 Hour', seconds: 3600 },
    { value: '6h', label: '6 Hours', seconds: 21600 },
    { value: '24h', label: '24 Hours', seconds: 86400 },
    { value: '7d', label: '7 Days', seconds: 604800 }
  ];

  // Chart type configurations
  const chartTypes = [
    { value: 'line', label: 'Line Chart', icon: faChartLine },
    { value: 'bar', label: 'Bar Chart', icon: faChartBar },
    { value: 'area', label: 'Area Chart', icon: faChartArea },
    { value: 'pie', label: 'Pie Chart', icon: faChartPie },
    { value: 'scatter', label: 'Scatter Plot', icon: faTh }
  ];

  // Data processing
  const processData = useCallback((rawData) => {
    if (!rawData || rawData.length === 0) return [];
    
    const range = timeRangeOptions.find(r => r.value === timeRange);
    const cutoffTime = Date.now() - (range.seconds * 1000);
    
    // Filter by time range
    let filteredData = rawData.filter(point => point.timestamp >= cutoffTime);
    
    // Apply smoothing if enabled
    if (smoothing && filteredData.length > 2) {
      filteredData = applyMovingAverage(filteredData, 3);
    }
    
    // Limit data points
    if (filteredData.length > maxDataPoints) {
      const step = Math.ceil(filteredData.length / maxDataPoints);
      filteredData = filteredData.filter((_, index) => index % step === 0);
    }
    
    return filteredData;
  }, [timeRange, smoothing, maxDataPoints]);

  const applyMovingAverage = (data, windowSize) => {
    return data.map((point, index) => {
      if (index < windowSize - 1) return point;
      
      const window = data.slice(index - windowSize + 1, index + 1);
      const smoothedValues = {};
      
      Object.keys(point.values || {}).forEach(key => {
        const sum = window.reduce((acc, p) => acc + (p.values[key] || 0), 0);
        smoothedValues[key] = sum / windowSize;
      });
      
      return { ...point, values: smoothedValues };
    });
  };

  // Statistics calculation
  const calculateStats = useCallback((data) => {
    if (!data || data.length === 0) return {};
    
    const stats = {};
    
    selectedMetrics.forEach(metricId => {
      const metric = metrics.find(m => m.id === metricId);
      if (!metric) return;
      
      const values = data.map(d => d.values?.[metricId] || 0).filter(v => !isNaN(v));
      
      if (values.length === 0) {
        stats[metricId] = { min: 0, max: 0, avg: 0, current: 0, trend: 'stable' };
        return;
      }
      
      const min = Math.min(...values);
      const max = Math.max(...values);
      const sum = values.reduce((acc, val) => acc + val, 0);
      const avg = sum / values.length;
      const current = values[values.length - 1];
      
      // Calculate trend
      let trend = 'stable';
      if (values.length >= 2) {
        const recent = values.slice(-Math.min(5, values.length));
        const older = values.slice(-Math.min(10, values.length), -Math.min(5, values.length));
        
        if (recent.length > 0 && older.length > 0) {
          const recentAvg = recent.reduce((acc, val) => acc + val, 0) / recent.length;
          const olderAvg = older.reduce((acc, val) => acc + val, 0) / older.length;
          
          const change = ((recentAvg - olderAvg) / olderAvg) * 100;
          trend = Math.abs(change) < 5 ? 'stable' : (change > 0 ? 'rising' : 'falling');
        }
      }
      
      stats[metricId] = { min, max, avg, current, trend };
    });
    
    return stats;
  }, [selectedMetrics, metrics]);

  // Alert checking
  const checkAlerts = useCallback((data, stats) => {
    const alerts = [];
    
    Object.entries(alertThresholds).forEach(([metricId, thresholds]) => {
      const stat = stats[metricId];
      const metric = metrics.find(m => m.id === metricId);
      
      if (!stat || !metric) return;
      
      if (thresholds.critical && stat.current >= thresholds.critical) {
        alerts.push({
          id: `${metricId}-critical`,
          severity: 'critical',
          metric: metric.name,
          value: stat.current,
          threshold: thresholds.critical,
          message: `${metric.name} has exceeded critical threshold`
        });
      } else if (thresholds.warning && stat.current >= thresholds.warning) {
        alerts.push({
          id: `${metricId}-warning`,
          severity: 'warning',
          metric: metric.name,
          value: stat.current,
          threshold: thresholds.warning,
          message: `${metric.name} has exceeded warning threshold`
        });
      }
    });
    
    setActiveAlerts(alerts);
  }, [alertThresholds, metrics]);

  // Real-time data updates
  useEffect(() => {
    if (!isRealtime || isPaused) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    intervalRef.current = setInterval(() => {
      if (onDataUpdate) {
        onDataUpdate();
      }
      
      // Simulate real-time data if no callback provided
      if (!onDataUpdate && data.length === 0) {
        const newPoint = {
          timestamp: Date.now(),
          values: {}
        };
        
        selectedMetrics.forEach(metricId => {
          const metric = metrics.find(m => m.id === metricId);
          if (metric) {
            // Generate realistic data based on metric type
            const baseValue = metric.baseValue || 50;
            const variance = metric.variance || 20;
            newPoint.values[metricId] = baseValue + (Math.random() - 0.5) * variance;
          }
        });
        
        dataBufferRef.current.push(newPoint);
        if (dataBufferRef.current.length > maxDataPoints * 2) {
          dataBufferRef.current = dataBufferRef.current.slice(-maxDataPoints);
        }
        
        const processed = processData(dataBufferRef.current);
        setChartData(processed);
        
        const newStats = calculateStats(processed);
        setStats(newStats);
        checkAlerts(processed, newStats);
      }
    }, updateInterval);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [isRealtime, isPaused, updateInterval, onDataUpdate, data, selectedMetrics, metrics, processData, calculateStats, checkAlerts, maxDataPoints]);

  // Update chart data when external data changes
  useEffect(() => {
    if (data && data.length > 0) {
      const processed = processData(data);
      setChartData(processed);
      
      const newStats = calculateStats(processed);
      setStats(newStats);
      checkAlerts(processed, newStats);
    }
  }, [data, processData, calculateStats, checkAlerts]);

  // Chart rendering
  const renderChart = () => {
    if (!chartData || chartData.length === 0) {
      return (
        <div className="chart-empty">
          <div className="text-center p-5">
            <FontAwesomeIcon icon={faChartLine} size="3x" className="text-muted mb-3" />
            <h5>No Data Available</h5>
            <p className="text-muted">Start collecting data to see visualizations here.</p>
            {!isRealtime && (
              <Button variant="primary" onClick={() => setIsRealtime(true)}>
                <FontAwesomeIcon icon={faPlay} className="me-2" />
                Start Real-time Updates
              </Button>
            )}
          </div>
        </div>
      );
    }

    // Chart would be rendered using a library like Chart.js, D3.js, or Recharts
    // This is a placeholder that shows the structure
    return (
      <div className="chart-container" style={{ height: isFullscreen ? '80vh' : height }}>
        <canvas
          ref={canvasRef}
          className="chart-canvas"
          width="100%"
          height="100%"
        />
        
        {/* Chart overlay for interactions */}
        <div className="chart-overlay">
          {showLegend && (
            <div className="chart-legend">
              {selectedMetrics.map((metricId, index) => {
                const metric = metrics.find(m => m.id === metricId);
                if (!metric) return null;
                
                const colors = colorSchemes[colorScheme];
                const color = colors[index % colors.length];
                
                return (
                  <div key={metricId} className="legend-item">
                    <span 
                      className="legend-color"
                      style={{ backgroundColor: color }}
                    />
                    <span className="legend-label">{metric.name}</span>
                    <span className="legend-value">
                      {stats[metricId]?.current?.toFixed(2) || '0'} {metric.unit}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
          
          {/* Chart controls overlay */}
          <div className="chart-controls">
            <div className="chart-status">
              {isRealtime && !isPaused && (
                <Badge bg="success" className="pulse">
                  <FontAwesomeIcon icon={faPlay} className="me-1" />
                  Live
                </Badge>
              )}
              {isPaused && (
                <Badge bg="warning">
                  <FontAwesomeIcon icon={faPause} className="me-1" />
                  Paused
                </Badge>
              )}
            </div>
          </div>
        </div>
      </div>
    );
  };

  // Metrics stats cards
  const renderMetricsStats = () => (
    <Row className="metrics-stats g-3">
      {selectedMetrics.slice(0, 4).map(metricId => {
        const metric = metrics.find(m => m.id === metricId);
        const stat = stats[metricId];
        
        if (!metric || !stat) return null;
        
        return (
          <Col key={metricId} sm={6} lg={3}>
            <Card className="metric-stat-card">
              <Card.Body className="p-3">
                <div className="metric-header">
                  <h6 className="metric-name">{metric.name}</h6>
                  <span className={`trend-indicator ${stat.trend}`}>
                    <FontAwesomeIcon 
                      icon={stat.trend === 'rising' ? faChartLine : 
                            stat.trend === 'falling' ? faChartLine : 
                            faChartLine} 
                      className={stat.trend === 'rising' ? 'text-success' :
                               stat.trend === 'falling' ? 'text-danger' :
                               'text-muted'}
                    />
                  </span>
                </div>
                <div className="metric-value">
                  <span className="current-value">
                    {stat.current.toFixed(2)}
                  </span>
                  <span className="unit">{metric.unit}</span>
                </div>
                <div className="metric-range">
                  <small className="text-muted">
                    Min: {stat.min.toFixed(1)} | 
                    Avg: {stat.avg.toFixed(1)} | 
                    Max: {stat.max.toFixed(1)}
                  </small>
                </div>
              </Card.Body>
            </Card>
          </Col>
        );
      })}
    </Row>
  );

  // Export functionality
  const handleExport = (format) => {
    const exportData = {
      title,
      chartType,
      timeRange,
      metrics: selectedMetrics.map(id => metrics.find(m => m.id === id)),
      data: chartData,
      stats,
      timestamp: new Date().toISOString()
    };

    if (onExport) {
      onExport(exportData, format);
    } else {
      // Default export implementation
      if (format === 'csv') {
        exportToCSV(exportData);
      } else if (format === 'json') {
        exportToJSON(exportData);
      } else if (format === 'image') {
        exportToImage();
      }
    }
  };

  const exportToCSV = (data) => {
    const headers = ['timestamp', ...data.metrics.map(m => m.name)];
    const csvContent = [
      headers.join(','),
      ...data.data.map(point => {
        return [
          new Date(point.timestamp).toISOString(),
          ...data.metrics.map(m => point.values[m.id] || 0)
        ].join(',');
      })
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title.replace(/\s+/g, '_')}_${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const exportToJSON = (data) => {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title.replace(/\s+/g, '_')}_${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const exportToImage = () => {
    if (canvasRef.current) {
      const link = document.createElement('a');
      link.download = `${title.replace(/\s+/g, '_')}_${Date.now()}.png`;
      link.href = canvasRef.current.toDataURL();
      link.click();
    }
  };

  return (
    <div className={`realtime-data-visualization ${className}`}>
      <style jsx>{`
        .realtime-data-visualization {
          background: var(--bg-primary);
          border-radius: var(--radius-card);
          border: 1px solid var(--border-primary);
          overflow: hidden;
        }
        
        .visualization-header {
          padding: 1rem 1.5rem;
          border-bottom: 1px solid var(--border-primary);
          background: var(--bg-subtle);
        }
        
        .header-title {
          margin: 0;
          color: var(--text-primary);
          font-weight: var(--font-weight-semibold);
        }
        
        .header-controls {
          display: flex;
          gap: 0.5rem;
          align-items: center;
          flex-wrap: wrap;
        }
        
        .control-group {
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        
        .chart-type-selector {
          display: flex;
          gap: 0.25rem;
        }
        
        .chart-type-btn {
          padding: 0.375rem 0.75rem;
          border: 1px solid var(--border-primary);
          background: var(--bg-primary);
          color: var(--text-secondary);
          border-radius: var(--radius-sm);
          transition: all var(--duration-fast);
        }
        
        .chart-type-btn:hover {
          background: var(--bg-secondary);
          color: var(--text-primary);
        }
        
        .chart-type-btn.active {
          background: var(--brand-primary);
          color: white;
          border-color: var(--brand-primary);
        }
        
        .metrics-stats {
          padding: 1rem 1.5rem;
          border-bottom: 1px solid var(--border-primary);
        }
        
        .metric-stat-card {
          border: 1px solid var(--border-primary);
          transition: all var(--duration-fast);
        }
        
        .metric-stat-card:hover {
          border-color: var(--brand-primary);
          box-shadow: var(--shadow-sm);
        }
        
        .metric-header {
          display: flex;
          justify-content: between;
          align-items: center;
          margin-bottom: 0.5rem;
        }
        
        .metric-name {
          margin: 0;
          font-size: 0.875rem;
          color: var(--text-secondary);
          font-weight: var(--font-weight-medium);
        }
        
        .trend-indicator {
          font-size: 0.75rem;
        }
        
        .metric-value {
          display: flex;
          align-items: baseline;
          gap: 0.25rem;
          margin-bottom: 0.5rem;
        }
        
        .current-value {
          font-size: 1.5rem;
          font-weight: var(--font-weight-bold);
          color: var(--text-primary);
        }
        
        .unit {
          font-size: 0.875rem;
          color: var(--text-tertiary);
        }
        
        .metric-range {
          font-size: 0.75rem;
        }
        
        .chart-container {
          position: relative;
          padding: 1rem;
        }
        
        .chart-canvas {
          width: 100%;
          height: 100%;
        }
        
        .chart-empty {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 100%;
          min-height: 300px;
        }
        
        .chart-overlay {
          position: absolute;
          top: 1rem;
          right: 1rem;
          z-index: 10;
        }
        
        .chart-legend {
          background: rgba(255, 255, 255, 0.95);
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-md);
          padding: 0.75rem;
          box-shadow: var(--shadow-sm);
          min-width: 200px;
        }
        
        .legend-item {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          margin-bottom: 0.5rem;
        }
        
        .legend-item:last-child {
          margin-bottom: 0;
        }
        
        .legend-color {
          width: 12px;
          height: 12px;
          border-radius: 2px;
        }
        
        .legend-label {
          flex: 1;
          font-size: 0.875rem;
          color: var(--text-primary);
        }
        
        .legend-value {
          font-size: 0.875rem;
          color: var(--text-secondary);
          font-weight: var(--font-weight-medium);
        }
        
        .chart-controls {
          position: absolute;
          top: 1rem;
          left: 1rem;
        }
        
        .chart-status .badge.pulse {
          animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
          0% { opacity: 1; }
          50% { opacity: 0.7; }
          100% { opacity: 1; }
        }
        
        .alerts-section {
          padding: 1rem 1.5rem;
          border-top: 1px solid var(--border-primary);
          background: var(--bg-subtle);
        }
        
        .alert-item {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          padding: 0.5rem;
          border-radius: var(--radius-sm);
          margin-bottom: 0.5rem;
        }
        
        .alert-item.critical {
          background: rgba(239, 68, 68, 0.1);
          border: 1px solid var(--error-light);
        }
        
        .alert-item.warning {
          background: rgba(245, 158, 11, 0.1);
          border: 1px solid var(--warning-light);
        }
        
        .fullscreen-container {
          position: fixed;
          top: 0;
          left: 0;
          width: 100vw;
          height: 100vh;
          background: var(--bg-primary);
          z-index: var(--z-modal);
        }
        
        .fullscreen-controls {
          position: absolute;
          top: 1rem;
          right: 1rem;
          z-index: var(--z-modal);
        }
        
        @media (max-width: 768px) {
          .header-controls {
            flex-direction: column;
            align-items: stretch;
            gap: 0.75rem;
          }
          
          .control-group {
            justify-content: space-between;
          }
          
          .chart-type-selector {
            justify-content: center;
          }
          
          .chart-overlay {
            position: static;
            margin-top: 1rem;
          }
          
          .chart-legend {
            background: transparent;
            border: none;
            box-shadow: none;
            padding: 0.5rem;
          }
        }
      `}</style>

      <div className="visualization-header">
        <div className="d-flex justify-content-between align-items-start">
          <h5 className="header-title">{title}</h5>
          
          <div className="header-controls">
            <div className="control-group">
              <div className="chart-type-selector">
                {chartTypes.map(type => (
                  <button
                    key={type.value}
                    className={`chart-type-btn ${chartType === type.value ? 'active' : ''}`}
                    onClick={() => setChartType(type.value)}
                    title={type.label}
                  >
                    <FontAwesomeIcon icon={type.icon} />
                  </button>
                ))}
              </div>
            </div>
            
            <div className="control-group">
              <Form.Select
                size="sm"
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value)}
                style={{ width: 'auto' }}
              >
                {timeRangeOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Form.Select>
            </div>
            
            <div className="control-group">
              {isRealtime ? (
                <Button
                  size="sm"
                  variant={isPaused ? "success" : "warning"}
                  onClick={() => setIsPaused(!isPaused)}
                >
                  <FontAwesomeIcon icon={isPaused ? faPlay : faPause} />
                </Button>
              ) : (
                <Button
                  size="sm"
                  variant="primary"
                  onClick={() => setIsRealtime(true)}
                >
                  <FontAwesomeIcon icon={faPlay} />
                </Button>
              )}
              
              <Button
                size="sm"
                variant="outline-secondary"
                onClick={() => setShowSettings(true)}
              >
                <FontAwesomeIcon icon={faCog} />
              </Button>
              
              {enableFullscreen && (
                <Button
                  size="sm"
                  variant="outline-secondary"
                  onClick={() => setIsFullscreen(!isFullscreen)}
                >
                  <FontAwesomeIcon icon={isFullscreen ? faCompress : faExpand} />
                </Button>
              )}
              
              {enableExport && (
                <Dropdown>
                  <Dropdown.Toggle size="sm" variant="outline-secondary">
                    <FontAwesomeIcon icon={faDownload} />
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => handleExport('csv')}>
                      Export as CSV
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => handleExport('json')}>
                      Export as JSON
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => handleExport('image')}>
                      Export as Image
                    </Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
              )}
            </div>
          </div>
        </div>
      </div>

      {Object.keys(stats).length > 0 && renderMetricsStats()}

      <div ref={chartContainerRef}>
        {renderChart()}
      </div>

      {activeAlerts.length > 0 && (
        <div className="alerts-section">
          <h6>Active Alerts</h6>
          {activeAlerts.map(alert => (
            <div key={alert.id} className={`alert-item ${alert.severity}`}>
              <FontAwesomeIcon 
                icon={alert.severity === 'critical' ? faExclamationCircle : faExclamationTriangle}
                className={alert.severity === 'critical' ? 'text-danger' : 'text-warning'}
              />
              <div>
                <strong>{alert.metric}</strong>: {alert.message}
                <br />
                <small>
                  Current: {alert.value.toFixed(2)} | Threshold: {alert.threshold}
                </small>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Settings Modal */}
      <Modal show={showSettings} onHide={() => setShowSettings(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Visualization Settings</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Color Scheme</Form.Label>
                <Form.Select 
                  value={colorScheme}
                  onChange={(e) => setColorScheme(e.target.value)}
                >
                  {Object.keys(colorSchemes).map(scheme => (
                    <option key={scheme} value={scheme}>
                      {scheme.charAt(0).toUpperCase() + scheme.slice(1)}
                    </option>
                  ))}
                </Form.Select>
              </Form.Group>
            </Col>
            
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Update Interval (ms)</Form.Label>
                <Form.Control
                  type="number"
                  value={updateInterval}
                  onChange={(e) => setUpdateInterval(Number(e.target.value))}
                  min="1000"
                  max="60000"
                  step="1000"
                />
              </Form.Group>
            </Col>
          </Row>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="switch"
              label="Show Legend"
              checked={showLegend}
              onChange={(e) => setShowLegend(e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="switch"
              label="Apply Smoothing"
              checked={smoothing}
              onChange={(e) => setSmoothing(e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>Visible Metrics</Form.Label>
            {metrics.map(metric => (
              <Form.Check
                key={metric.id}
                type="checkbox"
                label={`${metric.name} (${metric.unit})`}
                checked={selectedMetrics.includes(metric.id)}
                onChange={(e) => {
                  if (e.target.checked) {
                    setSelectedMetrics(prev => [...prev, metric.id]);
                  } else {
                    setSelectedMetrics(prev => prev.filter(id => id !== metric.id));
                  }
                }}
              />
            ))}
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowSettings(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>

      {isFullscreen && (
        <div className="fullscreen-container">
          <div className="fullscreen-controls">
            <Button
              variant="outline-light"
              onClick={() => setIsFullscreen(false)}
            >
              <FontAwesomeIcon icon={faCompress} />
            </Button>
          </div>
          {renderChart()}
        </div>
      )}
    </div>
  );
};

export default RealtimeDataVisualization;
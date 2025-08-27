import React, { useState, useEffect, useRef } from 'react';
import { Card, Badge } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faChartLine,
  faMicrochip,
  faMemory,
  faNetworkWired,
  faHardDrive,
  faTemperatureHigh,
  faWifi,
  faClock
} from '@fortawesome/free-solid-svg-icons';
import MetricsChart from './MetricsChart';

const RealTimeMetrics = ({ wsConnection, nodeId }) => {
  const [metrics, setMetrics] = useState({
    cpu: [],
    memory: [],
    network: [],
    disk: [],
    temperature: []
  });

  const [currentMetrics, setCurrentMetrics] = useState({
    cpu: 0,
    memory: 0,
    networkIn: 0,
    networkOut: 0,
    diskRead: 0,
    diskWrite: 0,
    temperature: 0,
    uptime: 0
  });

  const [connectionStatus, setConnectionStatus] = useState('disconnected');
  const maxDataPoints = 30;
  const updateInterval = useRef(null);

  useEffect(() => {
    if (wsConnection) {
      wsConnection.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'metrics' && data.nodeId === nodeId) {
            updateMetrics(data.metrics);
          }
        } catch (error) {
          console.error('Error parsing metrics data:', error);
        }
      };

      wsConnection.onopen = () => setConnectionStatus('connected');
      wsConnection.onclose = () => setConnectionStatus('disconnected');
      wsConnection.onerror = () => setConnectionStatus('error');
    }

    // Simulate real-time data if no WebSocket connection
    if (!wsConnection) {
      updateInterval.current = setInterval(() => {
        const simulatedMetrics = generateSimulatedMetrics();
        updateMetrics(simulatedMetrics);
      }, 2000);
    }

    return () => {
      if (updateInterval.current) {
        clearInterval(updateInterval.current);
      }
    };
  }, [wsConnection, nodeId]);

  const generateSimulatedMetrics = () => {
    return {
      cpu: Math.random() * 100,
      memory: Math.random() * 100,
      networkIn: Math.random() * 1000,
      networkOut: Math.random() * 1000,
      diskRead: Math.random() * 500,
      diskWrite: Math.random() * 500,
      temperature: 40 + Math.random() * 30,
      uptime: Date.now()
    };
  };

  const updateMetrics = (newMetrics) => {
    const timestamp = Date.now();
    
    setCurrentMetrics(newMetrics);

    setMetrics(prev => ({
      cpu: addDataPoint(prev.cpu, { timestamp, value: newMetrics.cpu }),
      memory: addDataPoint(prev.memory, { timestamp, value: newMetrics.memory }),
      network: addDataPoint(prev.network, { 
        timestamp, 
        value: (newMetrics.networkIn + newMetrics.networkOut) / 2 
      }),
      disk: addDataPoint(prev.disk, { 
        timestamp, 
        value: (newMetrics.diskRead + newMetrics.diskWrite) / 2 
      }),
      temperature: addDataPoint(prev.temperature, { 
        timestamp, 
        value: newMetrics.temperature 
      })
    }));
  };

  const addDataPoint = (dataArray, newPoint) => {
    const updated = [...dataArray, newPoint];
    return updated.slice(-maxDataPoints);
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatUptime = (uptime) => {
    const seconds = Math.floor((Date.now() - uptime) / 1000);
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'connected': return 'success';
      case 'disconnected': return 'secondary';
      case 'error': return 'danger';
      default: return 'warning';
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Real-Time Metrics</h2>
        <Badge bg={getStatusColor(connectionStatus)} className="d-flex align-items-center">
          <FontAwesomeIcon icon={faWifi} className="me-2" />
          {connectionStatus.charAt(0).toUpperCase() + connectionStatus.slice(1)}
        </Badge>
      </div>

      {/* Current Metrics Cards */}
      <div className="row mb-4">
        <div className="col-md-3">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faMicrochip} size="2x" className="text-primary mb-2" />
              <h4>{currentMetrics.cpu?.toFixed(1)}%</h4>
              <p className="mb-0 text-muted">CPU Usage</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-3">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faMemory} size="2x" className="text-info mb-2" />
              <h4>{currentMetrics.memory?.toFixed(1)}%</h4>
              <p className="mb-0 text-muted">Memory Usage</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-3">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faTemperatureHigh} size="2x" className="text-warning mb-2" />
              <h4>{currentMetrics.temperature?.toFixed(1)}°C</h4>
              <p className="mb-0 text-muted">Temperature</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-3">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faClock} size="2x" className="text-success mb-2" />
              <h4>{formatUptime(currentMetrics.uptime)}</h4>
              <p className="mb-0 text-muted">Uptime</p>
            </Card.Body>
          </Card>
        </div>
      </div>

      {/* Real-Time Charts */}
      <div className="row mb-4">
        <div className="col-md-6">
          <Card>
            <Card.Header>
              <h6 className="mb-0">
                <FontAwesomeIcon icon={faMicrochip} className="me-2" />
                CPU Usage History
              </h6>
            </Card.Header>
            <Card.Body>
              <MetricsChart 
                data={metrics.cpu}
                title="CPU Usage"
                color="#667eea"
              />
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-6">
          <Card>
            <Card.Header>
              <h6 className="mb-0">
                <FontAwesomeIcon icon={faMemory} className="me-2" />
                Memory Usage History
              </h6>
            </Card.Header>
            <Card.Body>
              <MetricsChart 
                data={metrics.memory}
                title="Memory Usage"
                color="#764ba2"
              />
            </Card.Body>
          </Card>
        </div>
      </div>

      <div className="row mb-4">
        <div className="col-md-6">
          <Card>
            <Card.Header>
              <h6 className="mb-0">
                <FontAwesomeIcon icon={faNetworkWired} className="me-2" />
                Network Activity
              </h6>
            </Card.Header>
            <Card.Body>
              <MetricsChart 
                data={metrics.network}
                title="Network I/O"
                color="#28a745"
              />
              <div className="mt-3 d-flex justify-content-between">
                <small className="text-muted">
                  In: {formatBytes(currentMetrics.networkIn)}/s
                </small>
                <small className="text-muted">
                  Out: {formatBytes(currentMetrics.networkOut)}/s
                </small>
              </div>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-6">
          <Card>
            <Card.Header>
              <h6 className="mb-0">
                <FontAwesomeIcon icon={faHardDrive} className="me-2" />
                Disk Activity
              </h6>
            </Card.Header>
            <Card.Body>
              <MetricsChart 
                data={metrics.disk}
                title="Disk I/O"
                color="#dc3545"
              />
              <div className="mt-3 d-flex justify-content-between">
                <small className="text-muted">
                  Read: {formatBytes(currentMetrics.diskRead)}/s
                </small>
                <small className="text-muted">
                  Write: {formatBytes(currentMetrics.diskWrite)}/s
                </small>
              </div>
            </Card.Body>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default RealTimeMetrics;
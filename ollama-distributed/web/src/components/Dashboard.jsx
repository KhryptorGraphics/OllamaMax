import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Badge, ProgressBar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faServer, 
  faCheckCircle, 
  faBolt, 
  faClock,
  faMicrochip,
  faMemory,
  faNetworkWired
} from '@fortawesome/free-solid-svg-icons';
import MetricsChart from './MetricsChart';

const Dashboard = ({ clusterStatus, nodes, metrics, realTimeMetrics }) => {
  const [systemMetrics, setSystemMetrics] = useState({
    totalNodes: 0,
    onlineNodes: 0,
    totalRequests: 0,
    avgLatency: 0,
    cpuUsage: 0,
    memoryUsage: 0,
    networkUsage: 0
  });

  useEffect(() => {
    if (nodes && clusterStatus && metrics) {
      const onlineNodes = nodes.filter(node => node.status === 'online').length;
      const totalRequests = metrics.totalRequests || 0;
      const avgLatency = metrics.avgLatency || 0;
      
      setSystemMetrics({
        totalNodes: nodes.length,
        onlineNodes: onlineNodes,
        totalRequests: totalRequests,
        avgLatency: avgLatency,
        cpuUsage: metrics.cpu_usage || 0,
        memoryUsage: metrics.memory_usage || 0,
        networkUsage: metrics.network_usage || 0
      });
    }
  }, [clusterStatus, nodes, metrics]);

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Enhanced Dashboard</h2>
        <div className="d-flex gap-2">
          <button className="btn btn-outline-primary btn-sm">
            <FontAwesomeIcon icon={faSync} className="me-2" />
            Refresh
          </button>
          <button className="btn btn-outline-success btn-sm">
            <FontAwesomeIcon icon={faDownload} className="me-2" />
            Export
          </button>
        </div>
      </div>
      
      <div className="metrics-grid">
        <div className="card metric-card">
          <div className="card-body text-center">
            <FontAwesomeIcon icon={faServer} size="2x" className="mb-2" />
            <h3>{systemMetrics.totalNodes}</h3>
            <p className="mb-0">Total Nodes</p>
          </div>
        </div>
        <div className="card metric-card">
          <div className="card-body text-center">
            <FontAwesomeIcon icon={faCheckCircle} size="2x" className="mb-2" />
            <h3>{systemMetrics.onlineNodes}</h3>
            <p className="mb-0">Online Nodes</p>
          </div>
        </div>
        <div className="card metric-card">
          <div className="card-body text-center">
            <FontAwesomeIcon icon={faBolt} size="2x" className="mb-2" />
            <h3>{systemMetrics.totalRequests}</h3>
            <p className="mb-0">Total Requests</p>
          </div>
        </div>
        <div className="card metric-card">
          <div className="card-body text-center">
            <FontAwesomeIcon icon={faClock} size="2x" className="mb-2" />
            <h3>{systemMetrics.avgLatency}ms</h3>
            <p className="mb-0">Avg Latency</p>
          </div>
        </div>
      </div>
      
      {/* Real-time Charts */}
      <div className="row mt-4">
        <div className="col-md-4">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faMicrochip} className="me-2" />
                CPU Usage
              </h5>
            </div>
            <div className="card-body">
              <div className="chart-container">
                <MetricsChart 
                  data={realTimeMetrics.cpu}
                  title="CPU Usage"
                  color="#667eea"
                />
              </div>
            </div>
          </div>
        </div>
        <div className="col-md-4">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faMemory} className="me-2" />
                Memory Usage
              </h5>
            </div>
            <div className="card-body">
              <div className="chart-container">
                <MetricsChart 
                  data={realTimeMetrics.memory}
                  title="Memory Usage"
                  color="#764ba2"
                />
              </div>
            </div>
          </div>
        </div>
        <div className="col-md-4">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faNetworkWired} className="me-2" />
                Network I/O
              </h5>
            </div>
            <div className="card-body">
              <div className="chart-container">
                <MetricsChart 
                  data={realTimeMetrics.network}
                  title="Network I/O"
                  color="#28a745"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div className="row mt-4">
        <div className="col-md-6">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">Cluster Status</h5>
            </div>
            <div className="card-body">
              <div className="mb-3">
                <strong>Node ID:</strong> 
                <div className="d-flex justify-content-between align-items-center">
                  <span className="ms-2 text-muted">
                    {clusterStatus?.node_id ? clusterStatus.node_id.substring(0, 8) + '...' : 'Unknown'}
                  </span>
                  {clusterStatus?.node_id && (
                    <button 
                      className="copy-button"
                      onClick={() => navigator.clipboard.writeText(clusterStatus.node_id)}
                      title="Copy Node ID"
                    >
                      <FontAwesomeIcon icon={faCopy} />
                    </button>
                  )}
                </div>
              </div>
              <div className="mb-3">
                <strong>Leader:</strong> 
                <div className="d-flex justify-content-between align-items-center">
                  <span className="ms-2">
                    {clusterStatus?.leader ? clusterStatus.leader.substring(0, 8) + '...' : 'Unknown'}
                  </span>
                  {clusterStatus?.leader && (
                    <button 
                      className="copy-button"
                      onClick={() => navigator.clipboard.writeText(clusterStatus.leader)}
                      title="Copy Leader ID"
                    >
                      <FontAwesomeIcon icon={faCopy} />
                    </button>
                  )}
                </div>
              </div>
              <div className="mb-3">
                <strong>Status:</strong> 
                <span className={`badge ms-2 ${clusterStatus?.status === 'healthy' ? 'bg-success' : 'bg-warning'}`}>
                  <FontAwesomeIcon 
                    icon={clusterStatus?.status === 'healthy' ? faHeart : faExclamationTriangle} 
                    className="me-1" 
                  />
                  {clusterStatus?.status || 'Unknown'}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div className="col-md-6">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">Node Overview</h5>
            </div>
            <div className="card-body">
              <div className="row">
                <div className="col-6">
                  <div className="text-center">
                    <div className="status-indicator status-online"></div>
                    <div>Online: {systemMetrics.onlineNodes}</div>
                  </div>
                </div>
                <div className="col-6">
                  <div className="text-center">
                    <div className="status-indicator status-offline"></div>
                    <div>Offline: {systemMetrics.totalNodes - systemMetrics.onlineNodes}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
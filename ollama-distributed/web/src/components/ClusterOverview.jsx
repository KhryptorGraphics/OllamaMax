import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Badge, ProgressBar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faServer, faCheckCircle, faBolt, faClock } from '@fortawesome/free-solid-svg-icons';

const ClusterOverview = ({ clusterStatus, nodes }) => {
  const [metrics, setMetrics] = useState({
    totalNodes: 0,
    onlineNodes: 0,
    totalRequests: 0,
    avgLatency: 0,
    cpuUsage: 0,
    memoryUsage: 0,
    networkUsage: 0
  });

  useEffect(() => {
    if (nodes && clusterStatus) {
      const onlineNodes = nodes.filter(node => node.status === 'online').length;
      const totalRequests = clusterStatus.metrics?.totalRequests || 0;
      const avgLatency = clusterStatus.metrics?.avgLatency || 0;
      
      setMetrics({
        totalNodes: nodes.length,
        onlineNodes: onlineNodes,
        totalRequests: totalRequests,
        avgLatency: avgLatency,
        cpuUsage: clusterStatus.metrics?.cpu_usage || 0,
        memoryUsage: clusterStatus.metrics?.memory_usage || 0,
        networkUsage: clusterStatus.metrics?.network_usage || 0
      });
    }
  }, [clusterStatus, nodes]);

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Cluster Overview</h2>
        <div className="d-flex gap-2">
          <button className="btn btn-outline-primary btn-sm">
            <FontAwesomeIcon icon={faServer} className="me-2" />
            Refresh
          </button>
          <button className="btn btn-outline-success btn-sm">
            <FontAwesomeIcon icon={faBolt} className="me-2" />
            Export
          </button>
        </div>
      </div>
      
      <Row className="metrics-grid mb-4">
        <Col md={3}>
          <Card className="metric-card text-center">
            <Card.Body>
              <FontAwesomeIcon icon={faServer} size="2x" className="mb-2" />
              <h3>{metrics.totalNodes}</h3>
              <p className="text-muted">Total Nodes</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="metric-card text-center">
            <Card.Body>
              <FontAwesomeIcon icon={faCheckCircle} size="2x" className="mb-2" />
              <h3>{metrics.onlineNodes}</h3>
              <p className="text-muted">Online Nodes</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="metric-card text-center">
            <Card.Body>
              <FontAwesomeIcon icon={faBolt} size="2x" className="mb-2" />
              <h3>{metrics.totalRequests}</h3>
              <p className="text-muted">Total Requests</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="metric-card text-center">
            <Card.Body>
              <FontAwesomeIcon icon={faClock} size="2x" className="mb-2" />
              <h3>{metrics.avgLatency}ms</h3>
              <p className="text-muted">Avg Latency</p>
            </Card.Body>
          </Card>
        </Col>
      </Row>
      
      <Row>
        <Col md={6}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faServer} className="me-2" />
                Resource Utilization
              </h5>
            </Card.Header>
            <Card.Body>
              <div className="mb-3">
                <div className="d-flex justify-content-between">
                  <span>CPU Usage</span>
                  <span>{metrics.cpuUsage.toFixed(1)}%</span>
                </div>
                <ProgressBar 
                  variant={metrics.cpuUsage > 80 ? 'danger' : metrics.cpuUsage > 60 ? 'warning' : 'success'} 
                  now={metrics.cpuUsage} 
                />
              </div>
              <div className="mb-3">
                <div className="d-flex justify-content-between">
                  <span>Memory Usage</span>
                  <span>{metrics.memoryUsage.toFixed(1)}%</span>
                </div>
                <ProgressBar 
                  variant={metrics.memoryUsage > 80 ? 'danger' : metrics.memoryUsage > 60 ? 'warning' : 'success'} 
                  now={metrics.memoryUsage} 
                />
              </div>
              <div className="mb-3">
                <div className="d-flex justify-content-between">
                  <span>Network Usage</span>
                  <span>{metrics.networkUsage.toFixed(1)}%</span>
                </div>
                <ProgressBar 
                  variant={metrics.networkUsage > 80 ? 'danger' : metrics.networkUsage > 60 ? 'warning' : 'info'} 
                  now={metrics.networkUsage} 
                />
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                Cluster Status
              </h5>
            </Card.Header>
            <Card.Body>
              <div className="mb-3">
                <strong>Node ID:</strong>
                <div className="d-flex justify-content-between align-items-center">
                  <code className="text-break">{clusterStatus?.node_id?.substring(0, 16)}...</code>
                  <button 
                    className="copy-button"
                    onClick={() => navigator.clipboard.writeText(clusterStatus?.node_id)}
                    title="Copy Node ID"
                  >
                    <FontAwesomeIcon icon={faCopy} />
                  </button>
                </div>
              </div>
              <div className="mb-3">
                <strong>Leader:</strong>
                <div className="d-flex justify-content-between align-items-center">
                  <code className="text-break">
                    {clusterStatus?.leader ? clusterStatus.leader.substring(0, 16) + '...' : 'Unknown'}
                  </code>
                  {clusterStatus?.leader && (
                    <button 
                      className="copy-button"
                      onClick={() => navigator.clipboard.writeText(clusterStatus?.leader)}
                      title="Copy Leader ID"
                    >
                      <FontAwesomeIcon icon={faCopy} />
                    </button>
                  )}
                </div>
              </div>
              <div className="mb-3">
                <strong>Status:</strong>
                <Badge 
                  bg={clusterStatus?.status === 'healthy' ? 'success' : 'warning'} 
                  className="ms-2"
                >
                  <FontAwesomeIcon 
                    icon={clusterStatus?.status === 'healthy' ? faHeart : faExclamationTriangle} 
                    className="me-1" 
                  />
                  {clusterStatus?.status || 'Unknown'}
                </Badge>
              </div>
              <div className="mb-3">
                <strong>Connected Peers:</strong>
                <Badge bg="info" className="ms-2">{clusterStatus?.peers || 0}</Badge>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ClusterOverview;
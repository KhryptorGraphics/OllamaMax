import React from 'react';
import { Card, Badge, ProgressBar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faServer, 
  faCopy, 
  faInfoCircle, 
  faPause, 
  faPowerOff,
  faCheck,
  faClock
} from '@fortawesome/free-solid-svg-icons';

const NodesView = ({ nodes, onCopy }) => {
  const onlineNodes = nodes.filter(node => node.status === 'online').length;
  const offlineNodes = nodes.filter(node => node.status === 'offline').length;

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Node Management</h2>
        <button className="btn btn-primary">
          <FontAwesomeIcon icon={faPlus} className="me-2" />
          Add Node
        </button>
      </div>
      
      <div className="row mb-4">
        <div className="col-md-6">
          <div className="card">
            <div className="card-body">
              <div className="row">
                <div className="col-6 text-center">
                  <div className="status-indicator status-online mb-2"></div>
                  <strong>Online: {onlineNodes}</strong>
                </div>
                <div className="col-6 text-center">
                  <div className="status-indicator status-offline mb-2"></div>
                  <strong>Offline: {offlineNodes}</strong>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div className="col-md-6">
          <div className="card">
            <div className="card-body">
              <div className="d-flex justify-content-between">
                <div>
                  <h5 className="mb-1">Total Nodes</h5>
                  <h3 className="mb-0">{nodes.length}</h3>
                </div>
                <div>
                  <h5 className="mb-1">Models</h5>
                  <h3 className="mb-0">
                    {nodes.reduce((total, node) => total + (node.models?.length || 0), 0)}
                  </h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div className="node-grid">
        {nodes.map(node => (
          <div key={node.id} className="node-card">
            <div className="d-flex justify-content-between align-items-center mb-3">
              <h5 className="mb-0">
                <span className={`status-indicator status-${node.status}`}></span>
                Node {node.id?.substring(0, 8)}...
              </h5>
              <button 
                className="copy-button"
                onClick={() => onCopy(node.id)}
                title="Copy Node ID"
              >
                <FontAwesomeIcon icon={faCopy} />
              </button>
            </div>
            
            <div className="mb-3">
              <small className="text-muted">Address:</small>
              <div className="d-flex justify-content-between align-items-center">
                <span className="font-monospace">{node.address}</span>
                <button 
                  className="copy-button"
                  onClick={() => onCopy(node.address)}
                  title="Copy Address"
                >
                  <FontAwesomeIcon icon={faCopy} />
                </button>
              </div>
            </div>
            
            {node.usage && (
              <div>
                <div className="mb-2">
                  <div className="d-flex justify-content-between">
                    <small>CPU Usage</small>
                    <small>{node.usage.cpu?.toFixed(1)}%</small>
                  </div>
                  <div className="progress">
                    <div 
                      className={`progress-bar ${
                        node.usage.cpu > 80 ? 'progress-bar-danger' : 
                        node.usage.cpu > 60 ? 'progress-bar-warning' : 
                        'progress-bar-success'
                      }`}
                      style={{ width: `${node.usage.cpu}%` }}
                    ></div>
                  </div>
                </div>
                
                <div className="mb-2">
                  <div className="d-flex justify-content-between">
                    <small>Memory Usage</small>
                    <small>{node.usage.memory?.toFixed(1)}%</small>
                  </div>
                  <div className="progress">
                    <div 
                      className={`progress-bar ${
                        node.usage.memory > 80 ? 'progress-bar-danger' : 
                        node.usage.memory > 60 ? 'progress-bar-warning' : 
                        'progress-bar-success'
                      }`}
                      style={{ width: `${node.usage.memory}%` }}
                    ></div>
                  </div>
                </div>
                
                <div className="mb-3">
                  <div className="d-flex justify-content-between">
                    <small>Bandwidth</small>
                    <small>{node.usage.bandwidth?.toFixed(1)} MB/s</small>
                  </div>
                  <div className="progress">
                    <div 
                      className="progress-bar"
                      style={{ 
                        width: `${Math.min((node.usage.bandwidth || 0) / 100 * 100, 100)}%`,
                        background: 'linear-gradient(45deg, #667eea, #764ba2)'
                      }}
                    ></div>
                  </div>
                </div>
              </div>
            )}
            
            <div className="mb-3">
              <small className="text-muted">Models: {node.models?.length || 0}</small>
            </div>
            
            <div className="mb-3">
              <span className={`badge ${
                node.status === 'ready' ? 'bg-success' : 
                node.status === 'loading' ? 'bg-warning' : 
                'bg-secondary'
              }`}>
                <FontAwesomeIcon 
                  icon={
                    node.status === 'ready' ? faCheck : 
                    node.status === 'loading' ? faSpinner : 
                    faClock
                  } 
                  className="me-1"
                />
                {
                  node.status === 'ready' ? 'Ready for Inference' : 
                  node.status === 'loading' ? 'Loading Models' : 
                  'Idle'
                }
              </span>
            </div>
            
            <div className="d-flex gap-2">
              <button className="btn btn-sm btn-outline-primary">
                <FontAwesomeIcon icon={faInfoCircle} className="me-1" />
                Details
              </button>
              <button className="btn btn-sm btn-outline-warning">
                <FontAwesomeIcon icon={faPause} className="me-1" />
                Drain
              </button>
              <button className="btn btn-sm btn-outline-danger">
                <FontAwesomeIcon icon={faPowerOff} className="me-1" />
                Shutdown
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default NodesView;
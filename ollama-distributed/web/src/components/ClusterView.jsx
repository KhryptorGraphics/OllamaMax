import React from 'react';
import { Card, Badge } from 'react-bootstrap';
import { FontAwesomeIcon } from './icons';
import { 
  faSitemap, 
  faCopy, 
  faSync, 
  faSignOutAlt,
  faCrown,
  faHeart,
  faExclamationTriangle,
  faUser
} from '@fortawesome/free-solid-svg-icons';

const ClusterView = ({ clusterStatus, onCopy }) => {
  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Cluster Management</h2>
        <div className="d-flex gap-2">
          <button className="btn btn-outline-primary">
            <FontAwesomeIcon icon={faSync} className="me-2" />
            Refresh
          </button>
          <button className="btn btn-outline-success">
            <FontAwesomeIcon icon={faChartLine} className="me-2" />
            Health Check
          </button>
        </div>
      </div>
      
      <div className="row">
        <div className="col-md-6">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faSitemap} className="me-2" />
                Cluster Information
              </h5>
            </div>
            <div className="card-body">
              <div className="mb-3">
                <strong>Current Node ID:</strong>
                <div className="d-flex justify-content-between align-items-center">
                  <code className="text-break">{clusterStatus.node_id}</code>
                  <button 
                    className="copy-button"
                    onClick={() => onCopy(clusterStatus.node_id)}
                    title="Copy Node ID"
                  >
                    <FontAwesomeIcon icon={faCopy} />
                  </button>
                </div>
              </div>
              <div className="mb-3">
                <strong>Leader Node:</strong>
                <div className="d-flex justify-content-between align-items-center">
                  <code className="text-break">{clusterStatus.leader || 'Unknown'}</code>
                  {clusterStatus.leader && (
                    <button 
                      className="copy-button"
                      onClick={() => onCopy(clusterStatus.leader)}
                      title="Copy Leader ID"
                    >
                      <FontAwesomeIcon icon={faCopy} />
                    </button>
                  )}
                </div>
              </div>
              <div className="mb-3">
                <strong>Is Leader:</strong> 
                <span className={`badge ms-2 ${clusterStatus.is_leader ? 'bg-success' : 'bg-secondary'}`}>
                  <FontAwesomeIcon 
                    icon={clusterStatus.is_leader ? faCrown : faUser} 
                    className="me-1" 
                  />
                  {clusterStatus.is_leader ? 'Yes' : 'No'}
                </span>
              </div>
              <div className="mb-3">
                <strong>Connected Peers:</strong> 
                <span className="badge bg-info ms-2">{clusterStatus.peers || 0}</span>
              </div>
              <div className="mb-3">
                <strong>Cluster Health:</strong> 
                <span className={`badge ms-2 ${
                  clusterStatus.status === 'healthy' ? 'bg-success' : 'bg-warning'
                }`}>
                  <FontAwesomeIcon 
                    icon={
                      clusterStatus.status === 'healthy' ? faHeart : faExclamationTriangle
                    } 
                    className="me-1" 
                  />
                  {clusterStatus.status || 'Unknown'}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div className="col-md-6">
          <div className="card">
            <div className="card-header">
              <h5 className="mb-0">
                <FontAwesomeIcon icon={faSitemap} className="me-2" />
                Cluster Actions
              </h5>
            </div>
            <div className="card-body">
              <div className="mb-3">
                <label className="form-label">Join Cluster</label>
                <div className="input-group">
                  <input 
                    type="text" 
                    className="form-control" 
                    placeholder="Enter cluster address"
                  />
                  <button className="btn btn-primary">
                    <FontAwesomeIcon icon={faPlusCircle} className="me-2" />
                    Join
                  </button>
                </div>
              </div>
              
              <div className="d-grid gap-2">
                <button className="btn btn-outline-warning">
                  <FontAwesomeIcon icon={faSignOutAlt} className="me-2" />
                  Leave Cluster
                </button>
                <button className="btn btn-outline-info">
                  <FontAwesomeIcon icon={faSync} className="me-2" />
                  Refresh Status
                </button>
                <button className="btn btn-outline-success">
                  <FontAwesomeIcon icon={faDownload} className="me-2" />
                  Export Configuration
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ClusterView;
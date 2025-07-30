import React from 'react';
import { Card, Badge, ProgressBar, Table } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faExchangeAlt, 
  faDownload, 
  faUpload, 
  faPause, 
  faTimes
} from '@fortawesome/free-solid-svg-icons';

const TransfersView = ({ transfers }) => {
  const activeTransfers = transfers.filter(t => t.status === 'active').length;
  const completedTransfers = transfers.filter(t => t.status === 'completed').length;
  const failedTransfers = transfers.filter(t => t.status === 'failed').length;

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Transfer Management</h2>
        <div className="d-flex gap-2">
          <span className="badge bg-primary">Active: {activeTransfers}</span>
          <span className="badge bg-success">Completed: {completedTransfers}</span>
          <span className="badge bg-danger">Failed: {failedTransfers}</span>
        </div>
      </div>
      
      <div className="table-container">
        <Table striped bordered hover>
          <thead>
            <tr>
              <th>Transfer ID</th>
              <th>Model</th>
              <th>Type</th>
              <th>Status</th>
              <th>Progress</th>
              <th>Speed</th>
              <th>ETA</th>
              <th>Peer</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {transfers.map(transfer => (
              <tr key={transfer.id}>
                <td>
                  <span className="font-monospace">{transfer.id?.substring(0, 8)}...</span>
                </td>
                <td>{transfer.model_name}</td>
                <td>
                  <span className={`badge ${
                    transfer.type === 'download' ? 'bg-primary' : 'bg-info'
                  }`}>
                    <FontAwesomeIcon 
                      icon={
                        transfer.type === 'download' ? faDownload : faUpload
                      } 
                      className="me-1"
                    />
                    {transfer.type}
                  </span>
                </td>
                <td>
                  <span className={`badge ${
                    transfer.status === 'completed' ? 'bg-success' : 
                    transfer.status === 'failed' ? 'bg-danger' : 'bg-primary'
                  }`}>
                    {transfer.status}
                  </span>
                </td>
                <td>
                  <div className="transfer-progress">
                    <div 
                      className="transfer-progress-bar"
                      style={{ width: `${transfer.progress || 0}%` }}
                    ></div>
                  </div>
                  <small>{(transfer.progress || 0).toFixed(1)}%</small>
                </td>
                <td>
                  {transfer.speed ? `${(transfer.speed / 1024 / 1024).toFixed(2)} MB/s` : 'N/A'}
                </td>
                <td>
                  {transfer.eta ? `${transfer.eta}s` : 'N/A'}
                </td>
                <td>
                  <span className="font-monospace">{transfer.peer_id?.substring(0, 8)}...</span>
                </td>
                <td>
                  <div className="d-flex gap-1">
                    <button 
                      className="btn btn-sm btn-outline-warning"
                      disabled={transfer.status === 'completed'}
                    >
                      <FontAwesomeIcon icon={faPause} />
                    </button>
                    <button 
                      className="btn btn-sm btn-outline-danger"
                      disabled={transfer.status === 'completed'}
                    >
                      <FontAwesomeIcon icon={faTimes} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      </div>
    </div>
  );
};

export default TransfersView;
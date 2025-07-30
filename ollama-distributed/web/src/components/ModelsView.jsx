import React from 'react';
import { Card, Badge, ProgressBar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faBrain, 
  faDownload, 
  faPlay, 
  faTrash,
  faCheck,
  faSpinner,
  faClock
} from '@fortawesome/free-solid-svg-icons';

const ModelsView = ({ models, onDownload, onDelete, autoDistribution, onToggleAutoDistribution }) => {
  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Model Management</h2>
        <button className="btn btn-primary">
          <FontAwesomeIcon icon={faPlus} className="me-2" />
          Add Model
        </button>
      </div>
      
      {/* Auto-distribution toggle */}
      <div className="auto-distribution-toggle">
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <h5 className="mb-1">Automatic Model Distribution</h5>
            <p className="text-muted mb-0">
              Automatically distribute models across nodes for optimal performance
            </p>
          </div>
          <label className="toggle-switch">
            <input 
              type="checkbox" 
              checked={autoDistribution}
              onChange={onToggleAutoDistribution}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>
      </div>
      
      <div className="model-grid">
        {models.map(model => (
          <div key={model.name} className="model-card">
            <div className="d-flex justify-content-between align-items-start mb-3">
              <h5 className="mb-0">{model.name}</h5>
              <span className={`badge ${
                model.status === 'available' ? 'bg-success' : 
                model.status === 'loading' ? 'bg-warning' : 
                'bg-secondary'
              }`}>
                <FontAwesomeIcon 
                  icon={
                    model.status === 'available' ? faCheck : 
                    model.status === 'loading' ? faSpinner : 
                    faClock
                  } 
                  className="me-1"
                />
                {model.status}
              </span>
            </div>
            
            <div className="mb-3">
              <small className="text-muted">Size:</small>
              <span className="ms-2">{model.size ? `${(model.size / 1024 / 1024).toFixed(2)} MB` : 'Unknown'}</span>
            </div>
            
            <div className="mb-3">
              <small className="text-muted">Replicas:</small>
              <span className="ms-2">{model.replicas?.length || 0}</span>
            </div>
            
            <div className="mb-3">
              <small className="text-muted">Inference Ready:</small>
              <span className="ms-2">
                <span className={`status-indicator status-${model.inference_ready ? 'ready' : 'loading'}`}></span>
                {model.inference_ready ? 'Yes' : 'No'}
              </span>
            </div>
            
            {model.distribution_progress && (
              <div className="mb-3">
                <small className="text-muted">Distribution Progress:</small>
                <div className="transfer-progress mt-1">
                  <div 
                    className="transfer-progress-bar"
                    style={{ width: `${model.distribution_progress}%` }}
                  ></div>
                </div>
                <small className="text-muted">{model.distribution_progress}%</small>
              </div>
            )}
            
            <div className="d-flex gap-2">
              <button 
                className="btn btn-sm btn-primary"
                onClick={() => onDownload(model.name)}
              >
                <FontAwesomeIcon icon={faDownload} className="me-1" />
                Download
              </button>
              <button 
                className="btn btn-sm btn-outline-success"
                disabled={!model.inference_ready}
              >
                <FontAwesomeIcon icon={faPlay} className="me-1" />
                Test
              </button>
              <button 
                className="btn btn-sm btn-outline-danger"
                onClick={() => onDelete(model.name)}
              >
                <FontAwesomeIcon icon={faTrash} className="me-1" />
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ModelsView;
import React, { useState } from 'react';
import { Card, Form, Button, Badge, Alert } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faCog,
  faSave,
  faUndo,
  faShield,
  faServer,
  faNetwork,
  faHardDrive,
  faExclamationTriangle
} from '@fortawesome/free-solid-svg-icons';

const SystemSettings = ({ settings = {}, onSave, onReset }) => {
  const [formData, setFormData] = useState({
    clusterName: settings.clusterName || 'Ollama Distributed Cluster',
    maxNodes: settings.maxNodes || 10,
    heartbeatInterval: settings.heartbeatInterval || 30,
    syncInterval: settings.syncInterval || 60,
    maxRetries: settings.maxRetries || 3,
    loadBalancingStrategy: settings.loadBalancingStrategy || 'round_robin',
    autoScaling: settings.autoScaling || false,
    enableMetrics: settings.enableMetrics || true,
    enableLogging: settings.enableLogging || true,
    logLevel: settings.logLevel || 'info',
    authenticationEnabled: settings.authenticationEnabled || false,
    sslEnabled: settings.sslEnabled || false,
    backupEnabled: settings.backupEnabled || true,
    backupInterval: settings.backupInterval || 24,
    ...settings
  });

  const [showAlert, setShowAlert] = useState(false);
  const [alertType, setAlertType] = useState('success');
  const [alertMessage, setAlertMessage] = useState('');

  const handleChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await onSave(formData);
      setAlertType('success');
      setAlertMessage('Settings saved successfully!');
      setShowAlert(true);
    } catch (error) {
      setAlertType('danger');
      setAlertMessage('Failed to save settings: ' + error.message);
      setShowAlert(true);
    }
  };

  const handleReset = () => {
    if (onReset) {
      onReset();
    }
    setFormData({
      clusterName: 'Ollama Distributed Cluster',
      maxNodes: 10,
      heartbeatInterval: 30,
      syncInterval: 60,
      maxRetries: 3,
      loadBalancingStrategy: 'round_robin',
      autoScaling: false,
      enableMetrics: true,
      enableLogging: true,
      logLevel: 'info',
      authenticationEnabled: false,
      sslEnabled: false,
      backupEnabled: true,
      backupInterval: 24
    });
    setAlertType('info');
    setAlertMessage('Settings reset to defaults');
    setShowAlert(true);
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>System Settings</h2>
        <div className="d-flex gap-2">
          <Button variant="outline-warning" onClick={handleReset}>
            <FontAwesomeIcon icon={faUndo} className="me-2" />
            Reset to Defaults
          </Button>
        </div>
      </div>

      {showAlert && (
        <Alert 
          variant={alertType} 
          dismissible 
          onClose={() => setShowAlert(false)}
          className="mb-4"
        >
          {alertMessage}
        </Alert>
      )}

      <Form onSubmit={handleSubmit}>
        <div className="row">
          {/* General Settings */}
          <div className="col-md-6">
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FontAwesomeIcon icon={faCog} className="me-2" />
                  General Settings
                </h5>
              </Card.Header>
              <Card.Body>
                <Form.Group className="mb-3">
                  <Form.Label>Cluster Name</Form.Label>
                  <Form.Control
                    type="text"
                    value={formData.clusterName}
                    onChange={(e) => handleChange('clusterName', e.target.value)}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Label>Maximum Nodes</Form.Label>
                  <Form.Control
                    type="number"
                    min="1"
                    max="100"
                    value={formData.maxNodes}
                    onChange={(e) => handleChange('maxNodes', parseInt(e.target.value))}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Label>Load Balancing Strategy</Form.Label>
                  <Form.Select
                    value={formData.loadBalancingStrategy}
                    onChange={(e) => handleChange('loadBalancingStrategy', e.target.value)}
                  >
                    <option value="round_robin">Round Robin</option>
                    <option value="least_connections">Least Connections</option>
                    <option value="weighted">Weighted</option>
                    <option value="random">Random</option>
                  </Form.Select>
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable Auto Scaling"
                    checked={formData.autoScaling}
                    onChange={(e) => handleChange('autoScaling', e.target.checked)}
                  />
                </Form.Group>
              </Card.Body>
            </Card>
          </div>

          {/* Network & Performance */}
          <div className="col-md-6">
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FontAwesomeIcon icon={faServer} className="me-2" />
                  Network & Performance
                </h5>
              </Card.Header>
              <Card.Body>
                <Form.Group className="mb-3">
                  <Form.Label>Heartbeat Interval (seconds)</Form.Label>
                  <Form.Control
                    type="number"
                    min="5"
                    max="300"
                    value={formData.heartbeatInterval}
                    onChange={(e) => handleChange('heartbeatInterval', parseInt(e.target.value))}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Label>Sync Interval (seconds)</Form.Label>
                  <Form.Control
                    type="number"
                    min="10"
                    max="3600"
                    value={formData.syncInterval}
                    onChange={(e) => handleChange('syncInterval', parseInt(e.target.value))}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Label>Max Retries</Form.Label>
                  <Form.Control
                    type="number"
                    min="1"
                    max="10"
                    value={formData.maxRetries}
                    onChange={(e) => handleChange('maxRetries', parseInt(e.target.value))}
                  />
                </Form.Group>
              </Card.Body>
            </Card>
          </div>

          {/* Security Settings */}
          <div className="col-md-6">
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FontAwesomeIcon icon={faShield} className="me-2" />
                  Security Settings
                </h5>
              </Card.Header>
              <Card.Body>
                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable Authentication"
                    checked={formData.authenticationEnabled}
                    onChange={(e) => handleChange('authenticationEnabled', e.target.checked)}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable SSL/TLS"
                    checked={formData.sslEnabled}
                    onChange={(e) => handleChange('sslEnabled', e.target.checked)}
                  />
                  {formData.sslEnabled && (
                    <Alert variant="warning" className="mt-2">
                      <FontAwesomeIcon icon={faExclamationTriangle} className="me-2" />
                      SSL certificates must be configured separately
                    </Alert>
                  )}
                </Form.Group>
              </Card.Body>
            </Card>
          </div>

          {/* Monitoring & Logging */}
          <div className="col-md-6">
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FontAwesomeIcon icon={faHardDrive} className="me-2" />
                  Monitoring & Logging
                </h5>
              </Card.Header>
              <Card.Body>
                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable Metrics Collection"
                    checked={formData.enableMetrics}
                    onChange={(e) => handleChange('enableMetrics', e.target.checked)}
                  />
                </Form.Group>

                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable System Logging"
                    checked={formData.enableLogging}
                    onChange={(e) => handleChange('enableLogging', e.target.checked)}
                  />
                </Form.Group>

                {formData.enableLogging && (
                  <Form.Group className="mb-3">
                    <Form.Label>Log Level</Form.Label>
                    <Form.Select
                      value={formData.logLevel}
                      onChange={(e) => handleChange('logLevel', e.target.value)}
                    >
                      <option value="debug">Debug</option>
                      <option value="info">Info</option>
                      <option value="warn">Warning</option>
                      <option value="error">Error</option>
                    </Form.Select>
                  </Form.Group>
                )}

                <Form.Group className="mb-3">
                  <Form.Check
                    type="checkbox"
                    label="Enable Automatic Backups"
                    checked={formData.backupEnabled}
                    onChange={(e) => handleChange('backupEnabled', e.target.checked)}
                  />
                </Form.Group>

                {formData.backupEnabled && (
                  <Form.Group className="mb-3">
                    <Form.Label>Backup Interval (hours)</Form.Label>
                    <Form.Control
                      type="number"
                      min="1"
                      max="168"
                      value={formData.backupInterval}
                      onChange={(e) => handleChange('backupInterval', parseInt(e.target.value))}
                    />
                  </Form.Group>
                )}
              </Card.Body>
            </Card>
          </div>
        </div>

        <div className="text-center mt-4">
          <Button variant="primary" type="submit" size="lg">
            <FontAwesomeIcon icon={faSave} className="me-2" />
            Save Settings
          </Button>
        </div>
      </Form>
    </div>
  );
};

export default SystemSettings;
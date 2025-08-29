import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Form, Modal, Table, Badge, ProgressBar, Alert } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faDownload,
  faTrash,
  faPlay,
  faStop,
  faCopy,
  faSearch,
  faFilter,
  faSort,
  faEye,
  faCog,
  faRocket,
  faDatabase,
  faCloudDownload,
  faHdd,
  faMicrochip,
  faMemory,
  faNetworkWired,
  faCheckCircle,
  faExclamationTriangle,
  faSpinner,
  faSyncAlt,
  faPlus,
  faEdit,
  faShareAlt
} from '@fortawesome/free-solid-svg-icons';
import LoadingSpinner from './LoadingSpinner';

const ModelManager = ({
  models = [],
  nodes = [],
  onModelDownload,
  onModelDelete,
  onModelDeploy,
  onModelStop,
  onModelReplicate,
  loading = false,
  error = null,
  className = ""
}) => {
  const [selectedModels, setSelectedModels] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [sortBy, setSortBy] = useState('name');
  const [sortOrder, setSortOrder] = useState('asc');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [selectedModel, setSelectedModel] = useState(null);
  const [deploymentSettings, setDeploymentSettings] = useState({});
  const [bulkAction, setBulkAction] = useState('');
  const [modelRegistry, setModelRegistry] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});

  // Available models from Ollama registry
  const availableModels = [
    { name: 'llama2:7b', size: '3.8GB', description: 'Meta\'s Llama 2 7B parameter model', category: 'General', popularity: 95 },
    { name: 'llama2:13b', size: '7.3GB', description: 'Meta\'s Llama 2 13B parameter model', category: 'General', popularity: 88 },
    { name: 'codellama:7b', size: '3.8GB', description: 'Code-specialized Llama model', category: 'Code', popularity: 92 },
    { name: 'codellama:13b', size: '7.3GB', description: 'Code-specialized Llama 13B model', category: 'Code', popularity: 85 },
    { name: 'mistral:7b', size: '4.1GB', description: 'Mistral AI\'s 7B parameter model', category: 'General', popularity: 89 },
    { name: 'neural-chat:7b', size: '4.1GB', description: 'Intel\'s neural chat model', category: 'Chat', popularity: 76 },
    { name: 'vicuna:7b', size: '3.8GB', description: 'Vicuna conversational AI model', category: 'Chat', popularity: 82 },
    { name: 'orca-mini:3b', size: '1.9GB', description: 'Microsoft\'s compact Orca model', category: 'General', popularity: 78 },
    { name: 'phi:2.7b', size: '1.7GB', description: 'Microsoft\'s Phi small language model', category: 'General', popularity: 73 },
    { name: 'stablelm:7b', size: '4.1GB', description: 'Stability AI\'s language model', category: 'General', popularity: 71 }
  ];

  useEffect(() => {
    setModelRegistry(availableModels);
  }, []);

  // Filter and sort models
  const filteredModels = models
    .filter(model => {
      const matchesSearch = model.name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesStatus = filterStatus === 'all' || model.status === filterStatus;
      return matchesSearch && matchesStatus;
    })
    .sort((a, b) => {
      let aVal = a[sortBy];
      let bVal = b[sortBy];
      
      if (sortBy === 'size') {
        aVal = a.size || 0;
        bVal = b.size || 0;
      }
      
      if (sortOrder === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });

  const handleModelSelect = (modelName) => {
    setSelectedModels(prev => 
      prev.includes(modelName) 
        ? prev.filter(m => m !== modelName)
        : [...prev, modelName]
    );
  };

  const handleSelectAll = () => {
    if (selectedModels.length === filteredModels.length) {
      setSelectedModels([]);
    } else {
      setSelectedModels(filteredModels.map(m => m.name));
    }
  };

  const handleBulkAction = async () => {
    if (!bulkAction || selectedModels.length === 0) return;

    try {
      switch (bulkAction) {
        case 'delete':
          if (window.confirm(`Delete ${selectedModels.length} models?`)) {
            for (const modelName of selectedModels) {
              await onModelDelete(modelName);
            }
          }
          break;
        case 'stop':
          for (const modelName of selectedModels) {
            await onModelStop(modelName);
          }
          break;
        case 'replicate':
          // Show replication settings modal
          break;
        default:
          break;
      }
      setSelectedModels([]);
      setBulkAction('');
    } catch (error) {
      console.error('Bulk action failed:', error);
    }
  };

  const getModelStatusBadge = (model) => {
    const statusConfig = {
      'available': { bg: 'success', icon: faCheckCircle, text: 'Available' },
      'downloading': { bg: 'info', icon: faSpinner, text: 'Downloading', spin: true },
      'error': { bg: 'danger', icon: faExclamationTriangle, text: 'Error' },
      'starting': { bg: 'warning', icon: faSpinner, text: 'Starting', spin: true },
      'running': { bg: 'success', icon: faPlay, text: 'Running' },
      'stopped': { bg: 'secondary', icon: faStop, text: 'Stopped' }
    };

    const config = statusConfig[model.status] || statusConfig['available'];
    
    return (
      <Badge bg={config.bg}>
        <FontAwesomeIcon 
          icon={config.icon} 
          className={`me-1 ${config.spin ? 'fa-spin' : ''}`} 
        />
        {config.text}
      </Badge>
    );
  };

  const formatSize = (bytes) => {
    if (typeof bytes === 'string') return bytes;
    if (!bytes) return 'Unknown';
    const gb = bytes / (1024 * 1024 * 1024);
    return `${gb.toFixed(1)} GB`;
  };

  const getModelMetrics = (model) => {
    // Mock metrics - in real app, these would come from monitoring
    return {
      requests: Math.floor(Math.random() * 1000),
      avgLatency: Math.floor(Math.random() * 300) + 50,
      errorRate: Math.random() * 5,
      uptime: Math.random() * 100
    };
  };

  const handleDownloadModel = async (modelName) => {
    try {
      setDownloadProgress(prev => ({ ...prev, [modelName]: 0 }));
      
      // Simulate download progress
      const interval = setInterval(() => {
        setDownloadProgress(prev => {
          const current = prev[modelName] || 0;
          if (current >= 100) {
            clearInterval(interval);
            return { ...prev, [modelName]: undefined };
          }
          return { ...prev, [modelName]: current + Math.random() * 10 };
        });
      }, 500);
      
      await onModelDownload(modelName);
    } catch (error) {
      setDownloadProgress(prev => ({ ...prev, [modelName]: undefined }));
      console.error('Download failed:', error);
    }
  };

  const renderModelCard = (model) => {
    const metrics = getModelMetrics(model);
    const isSelected = selectedModels.includes(model.name);
    const isDownloading = downloadProgress[model.name] !== undefined;
    
    return (
      <Card 
        key={model.name} 
        className={`model-card h-100 ${isSelected ? 'border-primary' : ''}`}
      >
        <Card.Header className="d-flex justify-content-between align-items-center">
          <div className="d-flex align-items-center">
            <Form.Check
              type="checkbox"
              checked={isSelected}
              onChange={() => handleModelSelect(model.name)}
              className="me-2"
            />
            <h6 className="mb-0">{model.name}</h6>
          </div>
          {getModelStatusBadge(model)}
        </Card.Header>
        
        <Card.Body>
          <div className="model-info mb-3">
            <div className="d-flex justify-content-between mb-2">
              <small className="text-muted">Size:</small>
              <small>{formatSize(model.size)}</small>
            </div>
            <div className="d-flex justify-content-between mb-2">
              <small className="text-muted">Replicas:</small>
              <small>{model.replicas?.length || 0}</small>
            </div>
            <div className="d-flex justify-content-between mb-2">
              <small className="text-muted">Requests:</small>
              <small>{metrics.requests.toLocaleString()}</small>
            </div>
            <div className="d-flex justify-content-between mb-2">
              <small className="text-muted">Avg Latency:</small>
              <small>{metrics.avgLatency}ms</small>
            </div>
          </div>
          
          {isDownloading && (
            <div className="mb-3">
              <div className="d-flex justify-content-between mb-1">
                <small>Downloading...</small>
                <small>{downloadProgress[model.name].toFixed(0)}%</small>
              </div>
              <ProgressBar 
                now={downloadProgress[model.name]} 
                size="sm" 
                animated 
                variant="info"
              />
            </div>
          )}
          
          <div className="model-actions">
            <div className="btn-group w-100" role="group">
              <Button 
                variant="outline-primary" 
                size="sm"
                onClick={() => {
                  setSelectedModel(model);
                  setShowDetailsModal(true);
                }}
                title="View Details"
              >
                <FontAwesomeIcon icon={faEye} />
              </Button>
              
              {model.status === 'available' ? (
                <Button 
                  variant="outline-success" 
                  size="sm"
                  onClick={() => onModelDeploy(model.name)}
                  title="Deploy"
                >
                  <FontAwesomeIcon icon={faRocket} />
                </Button>
              ) : model.status === 'running' ? (
                <Button 
                  variant="outline-warning" 
                  size="sm"
                  onClick={() => onModelStop(model.name)}
                  title="Stop"
                >
                  <FontAwesomeIcon icon={faStop} />
                </Button>
              ) : null}
              
              <Button 
                variant="outline-info" 
                size="sm"
                onClick={() => onModelReplicate(model.name)}
                title="Replicate"
              >
                <FontAwesomeIcon icon={faCopy} />
              </Button>
              
              <Button 
                variant="outline-danger" 
                size="sm"
                onClick={() => {
                  if (window.confirm(`Delete model ${model.name}?`)) {
                    onModelDelete(model.name);
                  }
                }}
                title="Delete"
              >
                <FontAwesomeIcon icon={faTrash} />
              </Button>
            </div>
          </div>
        </Card.Body>
      </Card>
    );
  };

  const renderModelRegistry = () => {
    return (
      <Modal show={showAddModal} onHide={() => setShowAddModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={faDownload} className="me-2" />
            Model Registry
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="mb-3">
            <Form.Control
              type="search"
              placeholder="Search available models..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          
          <div className="table-responsive" style={{ maxHeight: '400px' }}>
            <Table hover size="sm">
              <thead className="sticky-top bg-white">
                <tr>
                  <th>Model</th>
                  <th>Size</th>
                  <th>Category</th>
                  <th>Popularity</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {modelRegistry
                  .filter(model => 
                    model.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                    model.description.toLowerCase().includes(searchTerm.toLowerCase())
                  )
                  .map(model => {
                    const isInstalled = models.some(m => m.name === model.name);
                    const isDownloading = downloadProgress[model.name] !== undefined;
                    
                    return (
                      <tr key={model.name}>
                        <td>
                          <div>
                            <strong>{model.name}</strong>
                            <br />
                            <small className="text-muted">{model.description}</small>
                          </div>
                        </td>
                        <td>{model.size}</td>
                        <td>
                          <Badge bg="outline-secondary" size="sm">
                            {model.category}
                          </Badge>
                        </td>
                        <td>
                          <div className="d-flex align-items-center">
                            <div className="progress me-2" style={{ width: '60px', height: '6px' }}>
                              <div 
                                className="progress-bar bg-success" 
                                style={{ width: `${model.popularity}%` }}
                              ></div>
                            </div>
                            <small>{model.popularity}%</small>
                          </div>
                        </td>
                        <td>
                          {isDownloading ? (
                            <div className="d-flex align-items-center">
                              <FontAwesomeIcon icon={faSpinner} spin className="me-2" />
                              <small>{downloadProgress[model.name].toFixed(0)}%</small>
                            </div>
                          ) : isInstalled ? (
                            <Badge bg="success">Installed</Badge>
                          ) : (
                            <Button 
                              variant="outline-primary" 
                              size="sm"
                              onClick={() => handleDownloadModel(model.name)}
                            >
                              <FontAwesomeIcon icon={faDownload} className="me-1" />
                              Download
                            </Button>
                          )}
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </Table>
          </div>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowAddModal(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    );
  };

  const renderModelDetails = () => {
    if (!selectedModel) return null;
    
    const metrics = getModelMetrics(selectedModel);
    
    return (
      <Modal show={showDetailsModal} onHide={() => setShowDetailsModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={faDatabase} className="me-2" />
            {selectedModel.name}
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Row>
            <Col md={6}>
              <h6>Basic Information</h6>
              <Table size="sm" className="mb-4">
                <tbody>
                  <tr>
                    <td><strong>Name:</strong></td>
                    <td>{selectedModel.name}</td>
                  </tr>
                  <tr>
                    <td><strong>Size:</strong></td>
                    <td>{formatSize(selectedModel.size)}</td>
                  </tr>
                  <tr>
                    <td><strong>Status:</strong></td>
                    <td>{getModelStatusBadge(selectedModel)}</td>
                  </tr>
                  <tr>
                    <td><strong>Inference Ready:</strong></td>
                    <td>
                      <Badge bg={selectedModel.inference_ready ? 'success' : 'warning'}>
                        {selectedModel.inference_ready ? 'Yes' : 'No'}
                      </Badge>
                    </td>
                  </tr>
                </tbody>
              </Table>
            </Col>
            
            <Col md={6}>
              <h6>Performance Metrics</h6>
              <Table size="sm" className="mb-4">
                <tbody>
                  <tr>
                    <td><strong>Total Requests:</strong></td>
                    <td>{metrics.requests.toLocaleString()}</td>
                  </tr>
                  <tr>
                    <td><strong>Average Latency:</strong></td>
                    <td>{metrics.avgLatency}ms</td>
                  </tr>
                  <tr>
                    <td><strong>Error Rate:</strong></td>
                    <td>{metrics.errorRate.toFixed(2)}%</td>
                  </tr>
                  <tr>
                    <td><strong>Uptime:</strong></td>
                    <td>{metrics.uptime.toFixed(1)}%</td>
                  </tr>
                </tbody>
              </Table>
            </Col>
          </Row>
          
          {selectedModel.replicas && selectedModel.replicas.length > 0 && (
            <div>
              <h6>Replicas ({selectedModel.replicas.length})</h6>
              <div className="row">
                {selectedModel.replicas.map(nodeId => {
                  const node = nodes.find(n => n.id === nodeId);
                  return (
                    <div key={nodeId} className="col-md-6 mb-2">
                      <Card size="sm">
                        <Card.Body className="p-2">
                          <div className="d-flex justify-content-between align-items-center">
                            <span>
                              <FontAwesomeIcon icon={faServer} className="me-2" />
                              {nodeId}
                            </span>
                            <Badge bg={node?.status === 'online' ? 'success' : 'danger'}>
                              {node?.status || 'unknown'}
                            </Badge>
                          </div>
                          {node && (
                            <small className="text-muted d-block">
                              CPU: {node.usage?.cpu || 0}% | 
                              Memory: {node.usage?.memory || 0}%
                            </small>
                          )}
                        </Card.Body>
                      </Card>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <div className="d-flex justify-content-between w-100">
            <div>
              <Button 
                variant="outline-info" 
                size="sm"
                onClick={() => onModelReplicate(selectedModel.name)}
                className="me-2"
              >
                <FontAwesomeIcon icon={faCopy} className="me-1" />
                Replicate
              </Button>
              <Button 
                variant="outline-success" 
                size="sm"
                onClick={() => onModelDeploy(selectedModel.name)}
              >
                <FontAwesomeIcon icon={faRocket} className="me-1" />
                Deploy
              </Button>
            </div>
            <Button variant="secondary" onClick={() => setShowDetailsModal(false)}>
              Close
            </Button>
          </div>
        </Modal.Footer>
      </Modal>
    );
  };

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading model manager..." />;
  }

  if (error) {
    return (
      <Alert variant="danger">
        <Alert.Heading>Model Manager Error</Alert.Heading>
        <p>{error}</p>
      </Alert>
    );
  }

  return (
    <div className={`model-manager ${className}`}>
      {/* Header */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Model Manager</h2>
        <div className="d-flex align-items-center gap-2">
          <Button variant="primary" onClick={() => setShowAddModal(true)}>
            <FontAwesomeIcon icon={faPlus} className="me-2" />
            Add Model
          </Button>
        </div>
      </div>

      {/* Controls */}
      <Card className="mb-4">
        <Card.Body>
          <Row className="align-items-center">
            <Col md={3}>
              <Form.Control
                type="search"
                placeholder="Search models..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </Col>
            <Col md={2}>
              <Form.Select 
                value={filterStatus} 
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                <option value="all">All Status</option>
                <option value="available">Available</option>
                <option value="running">Running</option>
                <option value="downloading">Downloading</option>
                <option value="error">Error</option>
              </Form.Select>
            </Col>
            <Col md={2}>
              <Form.Select 
                value={sortBy} 
                onChange={(e) => setSortBy(e.target.value)}
              >
                <option value="name">Sort by Name</option>
                <option value="size">Sort by Size</option>
                <option value="status">Sort by Status</option>
              </Form.Select>
            </Col>
            <Col md={2}>
              <Button 
                variant="outline-secondary" 
                onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
              >
                <FontAwesomeIcon icon={faSort} className="me-1" />
                {sortOrder === 'asc' ? 'Ascending' : 'Descending'}
              </Button>
            </Col>
            <Col md={3}>
              {selectedModels.length > 0 && (
                <div className="d-flex gap-2">
                  <Form.Select 
                    value={bulkAction} 
                    onChange={(e) => setBulkAction(e.target.value)}
                    size="sm"
                  >
                    <option value="">Bulk Actions ({selectedModels.length})</option>
                    <option value="delete">Delete Selected</option>
                    <option value="stop">Stop Selected</option>
                    <option value="replicate">Replicate Selected</option>
                  </Form.Select>
                  <Button 
                    variant="outline-primary" 
                    size="sm"
                    onClick={handleBulkAction}
                    disabled={!bulkAction}
                  >
                    Apply
                  </Button>
                </div>
              )}
            </Col>
          </Row>
        </Card.Body>
      </Card>

      {/* Model Grid */}
      <div className="mb-3">
        <div className="d-flex justify-content-between align-items-center">
          <h6>
            Models ({filteredModels.length})
            {selectedModels.length > 0 && (
              <Badge bg="primary" className="ms-2">
                {selectedModels.length} selected
              </Badge>
            )}
          </h6>
          <Form.Check
            type="checkbox"
            label="Select All"
            checked={selectedModels.length === filteredModels.length && filteredModels.length > 0}
            onChange={handleSelectAll}
          />
        </div>
      </div>
      
      <Row>
        {filteredModels.map(model => (
          <Col key={model.name} lg={4} md={6} className="mb-4">
            {renderModelCard(model)}
          </Col>
        ))}
      </Row>

      {filteredModels.length === 0 && (
        <Card className="text-center">
          <Card.Body className="py-5">
            <FontAwesomeIcon icon={faDatabase} size="3x" className="text-muted mb-3" />
            <h5 className="text-muted">No models found</h5>
            <p className="text-muted">Try adjusting your search or filter criteria.</p>
            <Button variant="primary" onClick={() => setShowAddModal(true)}>
              <FontAwesomeIcon icon={faPlus} className="me-2" />
              Add Your First Model
            </Button>
          </Card.Body>
        </Card>
      )}

      {/* Modals */}
      {renderModelRegistry()}
      {renderModelDetails()}
    </div>
  );
};

export default ModelManager;
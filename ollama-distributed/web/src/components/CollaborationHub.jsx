import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Form, Modal, Table, Badge, Alert, Tabs, Tab } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faUsers,
  faComments,
  faShare,
  faCopy,
  faPlay,
  faStop,
  faEdit,
  faTrash,
  faPlus,
  faEye,
  faCog,
  faCode,
  faRocket,
  faHistory,
  faBookmark,
  faTag,
  faSearch,
  faFilter,
  faSort,
  faDownload,
  faUpload,
  faSyncAlt,
  faUserPlus,
  faUserMinus,
  faShieldAlt,
  faGlobe,
  faLock,
  faUserFriends,
  faProjectDiagram
} from '@fortawesome/free-solid-svg-icons';
import LoadingSpinner from './LoadingSpinner';

const CollaborationHub = ({
  projects = [],
  users = [],
  currentUser = null,
  onProjectCreate,
  onProjectUpdate,
  onProjectDelete,
  onProjectShare,
  onUserInvite,
  onUserRemove,
  loading = false,
  error = null,
  className = ""
}) => {
  const [activeTab, setActiveTab] = useState('projects');
  const [selectedProject, setSelectedProject] = useState(null);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [sortBy, setSortBy] = useState('updated');
  const [projectForm, setProjectForm] = useState({
    name: '',
    description: '',
    type: 'chat',
    visibility: 'private',
    model: 'llama2:7b',
    tags: [],
    collaborators: []
  });
  const [inviteForm, setInviteForm] = useState({
    email: '',
    role: 'collaborator',
    message: ''
  });
  const [shareSettings, setShareSettings] = useState({
    shareType: 'link',
    expiresIn: '7d',
    permissions: ['read']
  });

  // Mock data for demonstration
  const mockProjects = [
    {
      id: '1',
      name: 'AI Customer Support Bot',
      description: 'Intelligent customer service chatbot with context awareness',
      type: 'chat',
      model: 'llama2:7b',
      visibility: 'team',
      owner: 'admin',
      collaborators: ['alice', 'bob', 'charlie'],
      tags: ['customer-service', 'chatbot', 'production'],
      created: '2024-08-20T10:30:00Z',
      updated: '2024-08-24T15:45:00Z',
      status: 'active',
      conversations: 1247,
      rating: 4.7,
      shared: false
    },
    {
      id: '2',
      name: 'Code Review Assistant',
      description: 'AI-powered code review and suggestion system',
      type: 'code',
      model: 'codellama:13b',
      visibility: 'public',
      owner: 'alice',
      collaborators: ['admin', 'dave'],
      tags: ['code-review', 'development', 'automation'],
      created: '2024-08-18T09:15:00Z',
      updated: '2024-08-23T11:20:00Z',
      status: 'active',
      conversations: 856,
      rating: 4.5,
      shared: true
    },
    {
      id: '3',
      name: 'Documentation Generator',
      description: 'Automatically generate documentation from code and comments',
      type: 'document',
      model: 'llama2:13b',
      visibility: 'private',
      owner: 'bob',
      collaborators: ['alice'],
      tags: ['documentation', 'automation', 'development'],
      created: '2024-08-15T14:20:00Z',
      updated: '2024-08-22T16:30:00Z',
      status: 'draft',
      conversations: 234,
      rating: 4.2,
      shared: false
    }
  ];

  const mockCollaborators = [
    { id: 'admin', name: 'Admin User', email: 'admin@example.com', role: 'owner', avatar: null, status: 'online' },
    { id: 'alice', name: 'Alice Johnson', email: 'alice@example.com', role: 'collaborator', avatar: null, status: 'online' },
    { id: 'bob', name: 'Bob Smith', email: 'bob@example.com', role: 'collaborator', avatar: null, status: 'away' },
    { id: 'charlie', name: 'Charlie Brown', email: 'charlie@example.com', role: 'viewer', avatar: null, status: 'offline' },
    { id: 'dave', name: 'Dave Wilson', email: 'dave@example.com', role: 'collaborator', avatar: null, status: 'online' }
  ];

  useEffect(() => {
    // Initialize with mock data if no real data provided
    if (projects.length === 0) {
      // Would normally set the mock data here
    }
  }, [projects]);

  const projectTypes = [
    { value: 'chat', label: 'Chat Assistant', icon: faComments },
    { value: 'code', label: 'Code Helper', icon: faCode },
    { value: 'document', label: 'Documentation', icon: faEdit },
    { value: 'analysis', label: 'Data Analysis', icon: faProjectDiagram }
  ];

  const visibilityOptions = [
    { value: 'private', label: 'Private', icon: faLock, desc: 'Only you and invited collaborators' },
    { value: 'team', label: 'Team', icon: faUserFriends, desc: 'All team members can access' },
    { value: 'public', label: 'Public', icon: faGlobe, desc: 'Anyone can view and use' }
  ];

  const roleOptions = [
    { value: 'owner', label: 'Owner', desc: 'Full access and management' },
    { value: 'collaborator', label: 'Collaborator', desc: 'Can edit and contribute' },
    { value: 'viewer', label: 'Viewer', desc: 'Read-only access' }
  ];

  // Filter and sort projects
  const filteredProjects = (projects.length > 0 ? projects : mockProjects)
    .filter(project => {
      const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           project.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           project.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()));
      const matchesType = filterType === 'all' || project.type === filterType;
      return matchesSearch && matchesType;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'created':
          return new Date(b.created) - new Date(a.created);
        case 'updated':
          return new Date(b.updated) - new Date(a.updated);
        case 'rating':
          return b.rating - a.rating;
        default:
          return new Date(b.updated) - new Date(a.updated);
      }
    });

  const handleProjectSubmit = async (e) => {
    e.preventDefault();
    try {
      if (selectedProject) {
        await onProjectUpdate(selectedProject.id, projectForm);
      } else {
        await onProjectCreate(projectForm);
      }
      setShowProjectModal(false);
      resetProjectForm();
    } catch (error) {
      console.error('Project operation failed:', error);
    }
  };

  const handleProjectEdit = (project) => {
    setSelectedProject(project);
    setProjectForm({
      name: project.name,
      description: project.description,
      type: project.type,
      visibility: project.visibility,
      model: project.model,
      tags: project.tags,
      collaborators: project.collaborators
    });
    setShowProjectModal(true);
  };

  const handleProjectDelete = async (projectId) => {
    if (window.confirm('Are you sure you want to delete this project?')) {
      try {
        await onProjectDelete(projectId);
      } catch (error) {
        console.error('Project deletion failed:', error);
      }
    }
  };

  const handleProjectShare = (project) => {
    setSelectedProject(project);
    setShowShareModal(true);
  };

  const handleUserInvite = async (e) => {
    e.preventDefault();
    try {
      await onUserInvite(selectedProject.id, inviteForm);
      setShowInviteModal(false);
      setInviteForm({ email: '', role: 'collaborator', message: '' });
    } catch (error) {
      console.error('User invite failed:', error);
    }
  };

  const resetProjectForm = () => {
    setProjectForm({
      name: '',
      description: '',
      type: 'chat',
      visibility: 'private',
      model: 'llama2:7b',
      tags: [],
      collaborators: []
    });
    setSelectedProject(null);
  };

  const getStatusBadge = (status) => {
    const statusConfig = {
      'active': { bg: 'success', text: 'Active' },
      'draft': { bg: 'warning', text: 'Draft' },
      'archived': { bg: 'secondary', text: 'Archived' },
      'paused': { bg: 'info', text: 'Paused' }
    };
    const config = statusConfig[status] || statusConfig['active'];
    return <Badge bg={config.bg}>{config.text}</Badge>;
  };

  const getVisibilityIcon = (visibility) => {
    const visibilityIcons = {
      'private': faLock,
      'team': faUserFriends,
      'public': faGlobe
    };
    return visibilityIcons[visibility] || faLock;
  };

  const getUserStatus = (status) => {
    const statusColors = {
      'online': 'success',
      'away': 'warning',
      'offline': 'secondary'
    };
    return statusColors[status] || 'secondary';
  };

  const renderProjectCard = (project) => {
    const typeConfig = projectTypes.find(t => t.value === project.type);
    const collaboratorsData = mockCollaborators.filter(c => project.collaborators.includes(c.id));
    const isOwner = project.owner === (currentUser?.id || 'admin');
    
    return (
      <Card key={project.id} className="project-card h-100">
        <Card.Header className="d-flex justify-content-between align-items-start">
          <div className="d-flex align-items-center">
            <FontAwesomeIcon 
              icon={typeConfig?.icon || faProjectDiagram} 
              className="me-2 text-primary" 
            />
            <div>
              <h6 className="mb-1">{project.name}</h6>
              <div className="d-flex align-items-center gap-2">
                <FontAwesomeIcon 
                  icon={getVisibilityIcon(project.visibility)} 
                  className="text-muted" 
                  size="sm"
                />
                <small className="text-muted">{project.visibility}</small>
                {getStatusBadge(project.status)}
              </div>
            </div>
          </div>
          <div className="project-actions">
            <Button variant="outline-primary" size="sm" className="me-1">
              <FontAwesomeIcon icon={faEye} />
            </Button>
            {isOwner && (
              <Button 
                variant="outline-secondary" 
                size="sm" 
                className="me-1"
                onClick={() => handleProjectEdit(project)}
              >
                <FontAwesomeIcon icon={faEdit} />
              </Button>
            )}
            <Button 
              variant="outline-info" 
              size="sm"
              onClick={() => handleProjectShare(project)}
            >
              <FontAwesomeIcon icon={faShare} />
            </Button>
          </div>
        </Card.Header>
        
        <Card.Body>
          <p className="text-muted mb-3" style={{ fontSize: '0.9rem' }}>
            {project.description}
          </p>
          
          <div className="mb-3">
            <div className="d-flex justify-content-between align-items-center mb-2">
              <small className="text-muted">Model:</small>
              <Badge bg="outline-primary" size="sm">{project.model}</Badge>
            </div>
            <div className="d-flex justify-content-between align-items-center mb-2">
              <small className="text-muted">Conversations:</small>
              <small>{project.conversations?.toLocaleString() || 0}</small>
            </div>
            <div className="d-flex justify-content-between align-items-center mb-2">
              <small className="text-muted">Rating:</small>
              <div className="d-flex align-items-center">
                <small className="me-1">{'★'.repeat(Math.floor(project.rating || 0))}</small>
                <small>{project.rating?.toFixed(1) || 'N/A'}</small>
              </div>
            </div>
          </div>
          
          {project.tags && project.tags.length > 0 && (
            <div className="mb-3">
              <div className="d-flex flex-wrap gap-1">
                {project.tags.slice(0, 3).map(tag => (
                  <Badge key={tag} bg="outline-secondary" size="sm">
                    <FontAwesomeIcon icon={faTag} className="me-1" />
                    {tag}
                  </Badge>
                ))}
                {project.tags.length > 3 && (
                  <Badge bg="outline-secondary" size="sm">+{project.tags.length - 3}</Badge>
                )}
              </div>
            </div>
          )}
          
          <div className="collaborators-section">
            <div className="d-flex justify-content-between align-items-center mb-2">
              <small className="text-muted">Collaborators:</small>
              {isOwner && (
                <Button 
                  variant="outline-success" 
                  size="sm"
                  onClick={() => {
                    setSelectedProject(project);
                    setShowInviteModal(true);
                  }}
                >
                  <FontAwesomeIcon icon={faUserPlus} />
                </Button>
              )}
            </div>
            <div className="d-flex align-items-center">
              {collaboratorsData.slice(0, 4).map(collaborator => (
                <div key={collaborator.id} className="position-relative me-2">
                  <div 
                    className="collaborator-avatar"
                    style={{
                      width: '24px',
                      height: '24px',
                      borderRadius: '50%',
                      backgroundColor: '#ddd',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      fontWeight: 'bold'
                    }}
                    title={`${collaborator.name} (${collaborator.role})`}
                  >
                    {collaborator.name.charAt(0).toUpperCase()}
                  </div>
                  <div 
                    className="position-absolute bottom-0 end-0"
                    style={{
                      width: '8px',
                      height: '8px',
                      borderRadius: '50%',
                      border: '1px solid white'
                    }}
                    className={`bg-${getUserStatus(collaborator.status)}`}
                  ></div>
                </div>
              ))}
              {collaboratorsData.length > 4 && (
                <small className="text-muted">+{collaboratorsData.length - 4}</small>
              )}
            </div>
          </div>
        </Card.Body>
        
        <Card.Footer className="text-muted">
          <small>
            Updated {new Date(project.updated).toLocaleDateString()}
            {project.shared && (
              <Badge bg="info" size="sm" className="ms-2">
                <FontAwesomeIcon icon={faShare} className="me-1" />
                Shared
              </Badge>
            )}
          </small>
        </Card.Footer>
      </Card>
    );
  };

  const renderProjectModal = () => {
    return (
      <Modal show={showProjectModal} onHide={() => setShowProjectModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={selectedProject ? faEdit : faPlus} className="me-2" />
            {selectedProject ? 'Edit Project' : 'Create New Project'}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleProjectSubmit}>
          <Modal.Body>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Project Name</Form.Label>
                  <Form.Control
                    type="text"
                    value={projectForm.name}
                    onChange={(e) => setProjectForm({...projectForm, name: e.target.value})}
                    required
                    placeholder="Enter project name"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Project Type</Form.Label>
                  <Form.Select
                    value={projectForm.type}
                    onChange={(e) => setProjectForm({...projectForm, type: e.target.value})}
                  >
                    {projectTypes.map(type => (
                      <option key={type.value} value={type.value}>
                        {type.label}
                      </option>
                    ))}
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            
            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={projectForm.description}
                onChange={(e) => setProjectForm({...projectForm, description: e.target.value})}
                placeholder="Describe your project..."
              />
            </Form.Group>
            
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Model</Form.Label>
                  <Form.Select
                    value={projectForm.model}
                    onChange={(e) => setProjectForm({...projectForm, model: e.target.value})}
                  >
                    <option value="llama2:7b">Llama 2 7B</option>
                    <option value="llama2:13b">Llama 2 13B</option>
                    <option value="codellama:7b">Code Llama 7B</option>
                    <option value="codellama:13b">Code Llama 13B</option>
                    <option value="mistral:7b">Mistral 7B</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Visibility</Form.Label>
                  <Form.Select
                    value={projectForm.visibility}
                    onChange={(e) => setProjectForm({...projectForm, visibility: e.target.value})}
                  >
                    {visibilityOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label} - {option.desc}
                      </option>
                    ))}
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            
            <Form.Group className="mb-3">
              <Form.Label>Tags (comma-separated)</Form.Label>
              <Form.Control
                type="text"
                value={projectForm.tags.join(', ')}
                onChange={(e) => setProjectForm({
                  ...projectForm, 
                  tags: e.target.value.split(',').map(tag => tag.trim()).filter(tag => tag)
                })}
                placeholder="ai, chatbot, customer-service"
              />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowProjectModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              <FontAwesomeIcon icon={selectedProject ? faEdit : faPlus} className="me-1" />
              {selectedProject ? 'Update Project' : 'Create Project'}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    );
  };

  const renderShareModal = () => {
    return (
      <Modal show={showShareModal} onHide={() => setShowShareModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={faShare} className="me-2" />
            Share Project
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {selectedProject && (
            <div>
              <h6>{selectedProject.name}</h6>
              <p className="text-muted mb-4">{selectedProject.description}</p>
              
              <Form.Group className="mb-3">
                <Form.Label>Share Type</Form.Label>
                <Form.Select
                  value={shareSettings.shareType}
                  onChange={(e) => setShareSettings({...shareSettings, shareType: e.target.value})}
                >
                  <option value="link">Share Link</option>
                  <option value="embed">Embed Code</option>
                  <option value="api">API Access</option>
                </Form.Select>
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Expires In</Form.Label>
                <Form.Select
                  value={shareSettings.expiresIn}
                  onChange={(e) => setShareSettings({...shareSettings, expiresIn: e.target.value})}
                >
                  <option value="1d">1 Day</option>
                  <option value="7d">7 Days</option>
                  <option value="30d">30 Days</option>
                  <option value="never">Never</option>
                </Form.Select>
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Permissions</Form.Label>
                <div>
                  <Form.Check
                    type="checkbox"
                    label="Read Access"
                    checked={shareSettings.permissions.includes('read')}
                    onChange={(e) => {
                      const perms = shareSettings.permissions.filter(p => p !== 'read');
                      if (e.target.checked) perms.push('read');
                      setShareSettings({...shareSettings, permissions: perms});
                    }}
                  />
                  <Form.Check
                    type="checkbox"
                    label="Use/Chat Access"
                    checked={shareSettings.permissions.includes('use')}
                    onChange={(e) => {
                      const perms = shareSettings.permissions.filter(p => p !== 'use');
                      if (e.target.checked) perms.push('use');
                      setShareSettings({...shareSettings, permissions: perms});
                    }}
                  />
                </div>
              </Form.Group>
              
              <Alert variant="info">
                <FontAwesomeIcon icon={faShieldAlt} className="me-2" />
                Sharing will create a public link that allows others to access this project based on the permissions you've set.
              </Alert>
            </div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowShareModal(false)}>
            Cancel
          </Button>
          <Button variant="primary" onClick={() => {
            // Generate and copy share link
            navigator.clipboard.writeText(`https://ollama.example.com/shared/${selectedProject?.id}`);
            setShowShareModal(false);
          }}>
            <FontAwesomeIcon icon={faCopy} className="me-1" />
            Copy Share Link
          </Button>
        </Modal.Footer>
      </Modal>
    );
  };

  const renderInviteModal = () => {
    return (
      <Modal show={showInviteModal} onHide={() => setShowInviteModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={faUserPlus} className="me-2" />
            Invite Collaborator
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleUserInvite}>
          <Modal.Body>
            {selectedProject && (
              <div>
                <h6>{selectedProject.name}</h6>
                <p className="text-muted mb-4">Invite someone to collaborate on this project</p>
                
                <Form.Group className="mb-3">
                  <Form.Label>Email Address</Form.Label>
                  <Form.Control
                    type="email"
                    value={inviteForm.email}
                    onChange={(e) => setInviteForm({...inviteForm, email: e.target.value})}
                    required
                    placeholder="colleague@example.com"
                  />
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Role</Form.Label>
                  <Form.Select
                    value={inviteForm.role}
                    onChange={(e) => setInviteForm({...inviteForm, role: e.target.value})}
                  >
                    {roleOptions.filter(r => r.value !== 'owner').map(role => (
                      <option key={role.value} value={role.value}>
                        {role.label} - {role.desc}
                      </option>
                    ))}
                  </Form.Select>
                </Form.Group>
                
                <Form.Group className="mb-3">
                  <Form.Label>Personal Message (Optional)</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={2}
                    value={inviteForm.message}
                    onChange={(e) => setInviteForm({...inviteForm, message: e.target.value})}
                    placeholder="Add a personal message to your invitation..."
                  />
                </Form.Group>
              </div>
            )}
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowInviteModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              <FontAwesomeIcon icon={faUserPlus} className="me-1" />
              Send Invitation
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    );
  };

  const renderCollaboratorsTab = () => {
    return (
      <div>
        <div className="d-flex justify-content-between align-items-center mb-4">
          <h6>Team Members ({mockCollaborators.length})</h6>
          <Button variant="primary" size="sm" onClick={() => setShowInviteModal(true)}>
            <FontAwesomeIcon icon={faUserPlus} className="me-2" />
            Invite Member
          </Button>
        </div>
        
        <div className="table-responsive">
          <Table hover>
            <thead>
              <tr>
                <th>User</th>
                <th>Role</th>
                <th>Status</th>
                <th>Projects</th>
                <th>Last Active</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {mockCollaborators.map(user => (
                <tr key={user.id}>
                  <td>
                    <div className="d-flex align-items-center">
                      <div 
                        className="collaborator-avatar me-2"
                        style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '50%',
                          backgroundColor: '#ddd',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 'bold'
                        }}
                      >
                        {user.name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <div className="fw-medium">{user.name}</div>
                        <small className="text-muted">{user.email}</small>
                      </div>
                    </div>
                  </td>
                  <td>
                    <Badge bg={
                      user.role === 'owner' ? 'primary' :
                      user.role === 'collaborator' ? 'success' : 'info'
                    }>
                      {user.role}
                    </Badge>
                  </td>
                  <td>
                    <div className="d-flex align-items-center">
                      <div 
                        className={`status-dot me-2 bg-${getUserStatus(user.status)}`}
                        style={{
                          width: '8px',
                          height: '8px',
                          borderRadius: '50%'
                        }}
                      ></div>
                      {user.status}
                    </div>
                  </td>
                  <td>
                    {mockProjects.filter(p => p.collaborators.includes(user.id) || p.owner === user.id).length}
                  </td>
                  <td>
                    <small className="text-muted">2 hours ago</small>
                  </td>
                  <td>
                    {user.role !== 'owner' && (
                      <Button 
                        variant="outline-danger" 
                        size="sm"
                        onClick={() => onUserRemove(user.id)}
                      >
                        <FontAwesomeIcon icon={faUserMinus} />
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </div>
    );
  };

  if (loading) {
    return <LoadingSpinner size="xl" text="Loading collaboration hub..." />;
  }

  if (error) {
    return (
      <Alert variant="danger">
        <Alert.Heading>Collaboration Hub Error</Alert.Heading>
        <p>{error}</p>
      </Alert>
    );
  }

  return (
    <div className={`collaboration-hub ${className}`}>
      {/* Header */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Collaboration Hub</h2>
        <div className="d-flex align-items-center gap-2">
          <Button variant="primary" onClick={() => {
            resetProjectForm();
            setShowProjectModal(true);
          }}>
            <FontAwesomeIcon icon={faPlus} className="me-2" />
            New Project
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onSelect={setActiveTab} className="mb-4">
        <Tab eventKey="projects" title={
          <span>
            <FontAwesomeIcon icon={faProjectDiagram} className="me-2" />
            Projects ({filteredProjects.length})
          </span>
        }>
          {/* Project Controls */}
          <Card className="mb-4">
            <Card.Body>
              <Row className="align-items-center">
                <Col md={4}>
                  <Form.Control
                    type="search"
                    placeholder="Search projects..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </Col>
                <Col md={2}>
                  <Form.Select 
                    value={filterType} 
                    onChange={(e) => setFilterType(e.target.value)}
                  >
                    <option value="all">All Types</option>
                    {projectTypes.map(type => (
                      <option key={type.value} value={type.value}>
                        {type.label}
                      </option>
                    ))}
                  </Form.Select>
                </Col>
                <Col md={2}>
                  <Form.Select 
                    value={sortBy} 
                    onChange={(e) => setSortBy(e.target.value)}
                  >
                    <option value="updated">Last Updated</option>
                    <option value="created">Created Date</option>
                    <option value="name">Name</option>
                    <option value="rating">Rating</option>
                  </Form.Select>
                </Col>
                <Col md={4} className="d-flex justify-content-end gap-2">
                  <Button variant="outline-secondary" size="sm">
                    <FontAwesomeIcon icon={faDownload} className="me-1" />
                    Export
                  </Button>
                  <Button variant="outline-primary" size="sm">
                    <FontAwesomeIcon icon={faSyncAlt} className="me-1" />
                    Refresh
                  </Button>
                </Col>
              </Row>
            </Card.Body>
          </Card>

          {/* Projects Grid */}
          <Row>
            {filteredProjects.map(project => (
              <Col key={project.id} lg={4} md={6} className="mb-4">
                {renderProjectCard(project)}
              </Col>
            ))}
          </Row>

          {filteredProjects.length === 0 && (
            <Card className="text-center">
              <Card.Body className="py-5">
                <FontAwesomeIcon icon={faProjectDiagram} size="3x" className="text-muted mb-3" />
                <h5 className="text-muted">No projects found</h5>
                <p className="text-muted">Create your first collaborative project to get started.</p>
                <Button variant="primary" onClick={() => {
                  resetProjectForm();
                  setShowProjectModal(true);
                }}>
                  <FontAwesomeIcon icon={faPlus} className="me-2" />
                  Create Project
                </Button>
              </Card.Body>
            </Card>
          )}
        </Tab>

        <Tab eventKey="collaborators" title={
          <span>
            <FontAwesomeIcon icon={faUsers} className="me-2" />
            Team ({mockCollaborators.length})
          </span>
        }>
          {renderCollaboratorsTab()}
        </Tab>
      </Tabs>

      {/* Modals */}
      {renderProjectModal()}
      {renderShareModal()}
      {renderInviteModal()}
    </div>
  );
};

export default CollaborationHub;
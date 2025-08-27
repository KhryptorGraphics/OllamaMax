import React, { useState } from 'react';
import { Card, Badge, Modal, Form, Button } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faUser,
  faUserPlus,
  faUserEdit,
  faUserTimes,
  faShield,
  faKey,
  faEye,
  faEyeSlash
} from '@fortawesome/free-solid-svg-icons';

const UserManagement = ({ users = [], onAddUser, onEditUser, onDeleteUser }) => {
  const [showModal, setShowModal] = useState(false);
  const [editingUser, setEditingUser] = useState(null);
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    role: 'viewer',
    active: true
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    if (editingUser) {
      onEditUser(editingUser.id, formData);
    } else {
      onAddUser(formData);
    }
    resetForm();
  };

  const resetForm = () => {
    setFormData({
      username: '',
      email: '',
      password: '',
      role: 'viewer',
      active: true
    });
    setEditingUser(null);
    setShowModal(false);
    setShowPassword(false);
  };

  const handleEdit = (user) => {
    setEditingUser(user);
    setFormData({
      username: user.username,
      email: user.email,
      password: '',
      role: user.role,
      active: user.active
    });
    setShowModal(true);
  };

  const getRoleBadge = (role) => {
    const variants = {
      admin: 'danger',
      operator: 'warning',
      viewer: 'info'
    };
    return variants[role] || 'secondary';
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>User Management</h2>
        <Button variant="primary" onClick={() => setShowModal(true)}>
          <FontAwesomeIcon icon={faUserPlus} className="me-2" />
          Add User
        </Button>
      </div>

      <div className="user-grid">
        {users.map(user => (
          <Card key={user.id} className="user-card">
            <Card.Body>
              <div className="d-flex justify-content-between align-items-start mb-3">
                <div className="d-flex align-items-center">
                  <FontAwesomeIcon icon={faUser} size="2x" className="me-3 text-muted" />
                  <div>
                    <h5 className="mb-1">{user.username}</h5>
                    <small className="text-muted">{user.email}</small>
                  </div>
                </div>
                <Badge bg={user.active ? 'success' : 'secondary'}>
                  {user.active ? 'Active' : 'Inactive'}
                </Badge>
              </div>

              <div className="mb-3">
                <Badge bg={getRoleBadge(user.role)} className="me-2">
                  <FontAwesomeIcon icon={faShield} className="me-1" />
                  {user.role.charAt(0).toUpperCase() + user.role.slice(1)}
                </Badge>
                <small className="text-muted">Last login: {user.lastLogin || 'Never'}</small>
              </div>

              <div className="d-flex gap-2">
                <Button
                  variant="outline-primary"
                  size="sm"
                  onClick={() => handleEdit(user)}
                >
                  <FontAwesomeIcon icon={faUserEdit} className="me-1" />
                  Edit
                </Button>
                <Button
                  variant="outline-warning"
                  size="sm"
                  onClick={() => {/* Handle reset password */}}
                >
                  <FontAwesomeIcon icon={faKey} className="me-1" />
                  Reset Password
                </Button>
                <Button
                  variant="outline-danger"
                  size="sm"
                  onClick={() => onDeleteUser(user.id)}
                >
                  <FontAwesomeIcon icon={faUserTimes} className="me-1" />
                  Delete
                </Button>
              </div>
            </Card.Body>
          </Card>
        ))}
      </div>

      {/* Add/Edit User Modal */}
      <Modal show={showModal} onHide={resetForm}>
        <Modal.Header closeButton>
          <Modal.Title>
            <FontAwesomeIcon icon={editingUser ? faUserEdit : faUserPlus} className="me-2" />
            {editingUser ? 'Edit User' : 'Add New User'}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Username</Form.Label>
              <Form.Control
                type="text"
                value={formData.username}
                onChange={(e) => setFormData({...formData, username: e.target.value})}
                required
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Email</Form.Label>
              <Form.Control
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({...formData, email: e.target.value})}
                required
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Password</Form.Label>
              <div className="input-group">
                <Form.Control
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) => setFormData({...formData, password: e.target.value})}
                  required={!editingUser}
                  placeholder={editingUser ? 'Leave blank to keep current password' : 'Enter password'}
                />
                <Button
                  variant="outline-secondary"
                  onClick={() => setShowPassword(!showPassword)}
                  type="button"
                >
                  <FontAwesomeIcon icon={showPassword ? faEyeSlash : faEye} />
                </Button>
              </div>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Role</Form.Label>
              <Form.Select
                value={formData.role}
                onChange={(e) => setFormData({...formData, role: e.target.value})}
              >
                <option value="viewer">Viewer</option>
                <option value="operator">Operator</option>
                <option value="admin">Administrator</option>
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Check
                type="checkbox"
                label="Active user"
                checked={formData.active}
                onChange={(e) => setFormData({...formData, active: e.target.checked})}
              />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={resetForm}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              <FontAwesomeIcon icon={editingUser ? faUserEdit : faUserPlus} className="me-2" />
              {editingUser ? 'Update User' : 'Create User'}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;
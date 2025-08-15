/**
 * User Profile Component
 * 
 * Displays and allows editing of user profile information.
 */

import React, { useState } from 'react';
import { Button, Input, Card } from '../../design-system/index.js';
import { useAuth } from '../../contexts/AuthContext.jsx';
import { useTheme } from '../../design-system/theme/ThemeProvider.jsx';

const UserProfile = () => {
  const { user, updateProfile, changePassword, logout } = useAuth();
  const { theme } = useTheme();
  const [activeTab, setActiveTab] = useState('profile');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);

  // Profile form state
  const [profileData, setProfileData] = useState({
    firstName: user?.firstName || '',
    lastName: user?.lastName || '',
    email: user?.email || '',
    phone: user?.phone || '',
    company: user?.company || '',
    role: user?.role || ''
  });

  // Password form state
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  });

  // Handle profile form changes
  const handleProfileChange = (field) => (event) => {
    setProfileData(prev => ({
      ...prev,
      [field]: event.target.value
    }));
  };

  // Handle password form changes
  const handlePasswordChange = (field) => (event) => {
    setPasswordData(prev => ({
      ...prev,
      [field]: event.target.value
    }));
  };

  // Update profile
  const handleProfileUpdate = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setMessage(null);

    try {
      await updateProfile(profileData);
      setMessage('Profile updated successfully!');
    } catch (err) {
      setError(err.message || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  // Change password
  const handlePasswordChange = async (event) => {
    event.preventDefault();
    
    if (passwordData.newPassword !== passwordData.confirmPassword) {
      setError('New passwords do not match');
      return;
    }

    setLoading(true);
    setError(null);
    setMessage(null);

    try {
      await changePassword({
        currentPassword: passwordData.currentPassword,
        newPassword: passwordData.newPassword
      });
      setMessage('Password changed successfully!');
      setPasswordData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      });
    } catch (err) {
      setError(err.message || 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  // Container styles
  const containerStyles = {
    maxWidth: '800px',
    margin: '0 auto',
    padding: '2rem'
  };

  // Header styles
  const headerStyles = {
    marginBottom: '2rem'
  };

  const titleStyles = {
    fontSize: '2rem',
    fontWeight: 'bold',
    color: theme.colors.text,
    marginBottom: '0.5rem'
  };

  const subtitleStyles = {
    color: theme.colors.textSecondary,
    fontSize: '1rem'
  };

  // Tab styles
  const tabContainerStyles = {
    display: 'flex',
    borderBottom: `1px solid ${theme.colors.border}`,
    marginBottom: '2rem'
  };

  const tabStyles = (isActive) => ({
    padding: '1rem 1.5rem',
    border: 'none',
    backgroundColor: 'transparent',
    color: isActive ? theme.colors.primary : theme.colors.textSecondary,
    borderBottom: isActive ? `2px solid ${theme.colors.primary}` : '2px solid transparent',
    cursor: 'pointer',
    fontSize: '1rem',
    fontWeight: isActive ? '600' : '400',
    transition: 'all 0.2s ease'
  });

  // Message styles
  const messageStyles = (type) => ({
    padding: '0.75rem',
    borderRadius: '0.375rem',
    marginBottom: '1.5rem',
    fontSize: '0.875rem',
    backgroundColor: type === 'error' ? theme.colors.error + '10' : theme.colors.success + '10',
    border: `1px solid ${type === 'error' ? theme.colors.error : theme.colors.success}`,
    color: type === 'error' ? theme.colors.error : theme.colors.success
  });

  // Form styles
  const formStyles = {
    display: 'grid',
    gap: '1.5rem'
  };

  const formRowStyles = {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '1rem'
  };

  return (
    <div style={containerStyles}>
      <div style={headerStyles}>
        <h1 style={titleStyles}>User Profile</h1>
        <p style={subtitleStyles}>Manage your account settings and preferences</p>
      </div>

      {/* Tabs */}
      <div style={tabContainerStyles}>
        <button
          style={tabStyles(activeTab === 'profile')}
          onClick={() => setActiveTab('profile')}
        >
          Profile Information
        </button>
        <button
          style={tabStyles(activeTab === 'security')}
          onClick={() => setActiveTab('security')}
        >
          Security
        </button>
        <button
          style={tabStyles(activeTab === 'preferences')}
          onClick={() => setActiveTab('preferences')}
        >
          Preferences
        </button>
      </div>

      {/* Messages */}
      {message && (
        <div style={messageStyles('success')}>
          {message}
        </div>
      )}

      {error && (
        <div style={messageStyles('error')}>
          {error}
        </div>
      )}

      {/* Profile Tab */}
      {activeTab === 'profile' && (
        <Card variant="elevated" size="lg">
          <Card.Header>
            <Card.Title>Profile Information</Card.Title>
            <Card.Description>
              Update your personal information and contact details
            </Card.Description>
          </Card.Header>

          <Card.Body>
            <form onSubmit={handleProfileUpdate} style={formStyles}>
              <div style={formRowStyles}>
                <Input
                  label="First Name"
                  value={profileData.firstName}
                  onChange={handleProfileChange('firstName')}
                  required
                />
                <Input
                  label="Last Name"
                  value={profileData.lastName}
                  onChange={handleProfileChange('lastName')}
                  required
                />
              </div>

              <Input
                type="email"
                label="Email Address"
                value={profileData.email}
                onChange={handleProfileChange('email')}
                required
              />

              <div style={formRowStyles}>
                <Input
                  type="tel"
                  label="Phone Number"
                  value={profileData.phone}
                  onChange={handleProfileChange('phone')}
                  placeholder="+1 (555) 123-4567"
                />
                <Input
                  label="Company"
                  value={profileData.company}
                  onChange={handleProfileChange('company')}
                  placeholder="Your company name"
                />
              </div>

              <Input
                label="Role"
                value={profileData.role}
                onChange={handleProfileChange('role')}
                placeholder="Your job title"
              />

              <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setProfileData({
                    firstName: user?.firstName || '',
                    lastName: user?.lastName || '',
                    email: user?.email || '',
                    phone: user?.phone || '',
                    company: user?.company || '',
                    role: user?.role || ''
                  })}
                >
                  Reset
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  loading={loading}
                  disabled={loading}
                >
                  Save Changes
                </Button>
              </div>
            </form>
          </Card.Body>
        </Card>
      )}

      {/* Security Tab */}
      {activeTab === 'security' && (
        <div style={{ display: 'grid', gap: '2rem' }}>
          <Card variant="elevated" size="lg">
            <Card.Header>
              <Card.Title>Change Password</Card.Title>
              <Card.Description>
                Update your password to keep your account secure
              </Card.Description>
            </Card.Header>

            <Card.Body>
              <form onSubmit={handlePasswordChange} style={formStyles}>
                <Input
                  type="password"
                  label="Current Password"
                  value={passwordData.currentPassword}
                  onChange={handlePasswordChange('currentPassword')}
                  required
                  autoComplete="current-password"
                />

                <Input
                  type="password"
                  label="New Password"
                  value={passwordData.newPassword}
                  onChange={handlePasswordChange('newPassword')}
                  required
                  autoComplete="new-password"
                  helperText="Password must be at least 8 characters long"
                />

                <Input
                  type="password"
                  label="Confirm New Password"
                  value={passwordData.confirmPassword}
                  onChange={handlePasswordChange('confirmPassword')}
                  required
                  autoComplete="new-password"
                />

                <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setPasswordData({
                      currentPassword: '',
                      newPassword: '',
                      confirmPassword: ''
                    })}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    variant="primary"
                    loading={loading}
                    disabled={loading}
                  >
                    Change Password
                  </Button>
                </div>
              </form>
            </Card.Body>
          </Card>

          <Card variant="elevated" size="lg">
            <Card.Header>
              <Card.Title>Account Actions</Card.Title>
              <Card.Description>
                Manage your account security and access
              </Card.Description>
            </Card.Header>

            <Card.Body>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <Button
                  variant="danger"
                  onClick={logout}
                >
                  Sign Out
                </Button>
              </div>
            </Card.Body>
          </Card>
        </div>
      )}

      {/* Preferences Tab */}
      {activeTab === 'preferences' && (
        <Card variant="elevated" size="lg">
          <Card.Header>
            <Card.Title>Preferences</Card.Title>
            <Card.Description>
              Customize your experience and notification settings
            </Card.Description>
          </Card.Header>

          <Card.Body>
            <div style={{ textAlign: 'center', padding: '2rem', color: theme.colors.textSecondary }}>
              <p>Preferences settings will be available in the next iteration.</p>
            </div>
          </Card.Body>
        </Card>
      )}
    </div>
  );
};

export default UserProfile;

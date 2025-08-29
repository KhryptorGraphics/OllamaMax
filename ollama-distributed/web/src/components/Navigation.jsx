import React, { useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faBars,
  faTimes,
  faHome,
  faServer,
  faCube,
  faExchangeAlt,
  faNetworkWired,
  faChartLine,
  faUsers,
  faDatabase,
  faCog,
  faSignOutAlt,
  faUser,
  faBell,
  faSearch
} from '@fortawesome/free-solid-svg-icons';
import ThemeToggle from './ThemeToggle';
import '../styles/theme.css';

const Navigation = ({ activeTab, onTabChange, user, onLogout }) => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [profileOpen, setProfileOpen] = useState(false);
  const [notificationOpen, setNotificationOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: faHome, badge: null },
    { id: 'nodes', label: 'Nodes', icon: faServer, badge: 3 },
    { id: 'models', label: 'Models', icon: faCube, badge: null },
    { id: 'transfers', label: 'Transfers', icon: faExchangeAlt, badge: 2 },
    { id: 'cluster', label: 'Cluster', icon: faNetworkWired, badge: null },
    { id: 'analytics', label: 'Analytics', icon: faChartLine, badge: null },
    { id: 'users', label: 'Users', icon: faUsers, badge: null },
    { id: 'database', label: 'Database', icon: faDatabase, badge: null },
    { id: 'settings', label: 'Settings', icon: faCog, badge: null }
  ];

  const notifications = [
    { id: 1, type: 'success', message: 'Model deployed successfully', time: '5 mins ago' },
    { id: 2, type: 'warning', message: 'Node 2 high memory usage', time: '10 mins ago' },
    { id: 3, type: 'info', message: 'New model available', time: '1 hour ago' }
  ];

  return (
    <>
      {/* Top Navigation Bar */}
      <nav className="navbar">
        <div className="navbar-container">
          <div className="navbar-left">
            <button
              className="btn-icon navbar-toggle"
              onClick={() => setSidebarOpen(!sidebarOpen)}
            >
              <FontAwesomeIcon icon={sidebarOpen ? faTimes : faBars} />
            </button>
            <div className="navbar-brand">
              <span className="brand-text">OllamaMax</span>
              <span className="brand-tag">Distributed AI</span>
            </div>
          </div>

          <div className="navbar-center">
            <div className="search-box">
              <FontAwesomeIcon icon={faSearch} className="search-icon" />
              <input
                type="text"
                className="search-input"
                placeholder="Search models, nodes, or settings..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>

          <div className="navbar-right">
            <ThemeToggle />
            
            <div className="notification-dropdown">
              <button
                className="btn-icon notification-btn"
                onClick={() => setNotificationOpen(!notificationOpen)}
              >
                <FontAwesomeIcon icon={faBell} />
                <span className="notification-badge">3</span>
              </button>
              
              {notificationOpen && (
                <div className="dropdown-menu notification-menu animate-fadeIn">
                  <div className="dropdown-header">
                    <h4>Notifications</h4>
                    <a href="#" className="text-muted">Mark all read</a>
                  </div>
                  <div className="dropdown-body">
                    {notifications.map(notif => (
                      <div key={notif.id} className={`notification-item ${notif.type}`}>
                        <div className="notification-content">
                          <p>{notif.message}</p>
                          <span className="notification-time">{notif.time}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="dropdown-footer">
                    <a href="#">View all notifications</a>
                  </div>
                </div>
              )}
            </div>

            <div className="profile-dropdown">
              <button
                className="profile-btn"
                onClick={() => setProfileOpen(!profileOpen)}
              >
                <div className="profile-avatar">
                  <FontAwesomeIcon icon={faUser} />
                </div>
                <span className="profile-name">{user?.name || 'Admin'}</span>
              </button>
              
              {profileOpen && (
                <div className="dropdown-menu profile-menu animate-fadeIn">
                  <div className="dropdown-header">
                    <div className="profile-info">
                      <strong>{user?.name || 'Administrator'}</strong>
                      <span>{user?.email || 'admin@ollamamax.io'}</span>
                    </div>
                  </div>
                  <div className="dropdown-body">
                    <a href="#" className="dropdown-item">
                      <FontAwesomeIcon icon={faUser} /> Profile
                    </a>
                    <a href="#" className="dropdown-item">
                      <FontAwesomeIcon icon={faCog} /> Preferences
                    </a>
                  </div>
                  <div className="dropdown-footer">
                    <button className="btn btn-danger btn-sm" onClick={onLogout}>
                      <FontAwesomeIcon icon={faSignOutAlt} /> Sign Out
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </nav>

      {/* Sidebar Navigation */}
      <aside className={`sidebar ${sidebarOpen ? 'sidebar-open' : 'sidebar-closed'}`}>
        <div className="sidebar-content">
          <nav className="sidebar-nav">
            {menuItems.map(item => (
              <button
                key={item.id}
                className={`sidebar-item ${activeTab === item.id ? 'active' : ''}`}
                onClick={() => onTabChange(item.id)}
              >
                <FontAwesomeIcon icon={item.icon} className="sidebar-icon" />
                {sidebarOpen && (
                  <>
                    <span className="sidebar-label">{item.label}</span>
                    {item.badge && (
                      <span className="sidebar-badge">{item.badge}</span>
                    )}
                  </>
                )}
              </button>
            ))}
          </nav>
        </div>
      </aside>

      <style jsx>{`
        .navbar {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          height: 64px;
          background: var(--background);
          border-bottom: 1px solid var(--border-color);
          z-index: var(--z-sticky);
        }

        .navbar-container {
          height: 100%;
          padding: 0 var(--spacing-lg);
          display: flex;
          align-items: center;
          justify-content: space-between;
        }

        .navbar-left,
        .navbar-right {
          display: flex;
          align-items: center;
          gap: var(--spacing-md);
        }

        .navbar-toggle {
          display: none;
        }

        .navbar-brand {
          display: flex;
          flex-direction: column;
          line-height: 1.2;
        }

        .brand-text {
          font-size: var(--font-size-xl);
          font-weight: var(--font-weight-bold);
          color: var(--primary-color);
        }

        .brand-tag {
          font-size: var(--font-size-xs);
          color: var(--text-muted);
        }

        .search-box {
          position: relative;
          width: 400px;
        }

        .search-icon {
          position: absolute;
          left: var(--spacing-md);
          top: 50%;
          transform: translateY(-50%);
          color: var(--text-muted);
        }

        .search-input {
          width: 100%;
          padding: var(--spacing-sm) var(--spacing-md);
          padding-left: calc(var(--spacing-xl) + var(--spacing-sm));
          border: 1px solid var(--border-color);
          border-radius: var(--radius-full);
          background: var(--surface);
          color: var(--text-primary);
          font-size: var(--font-size-sm);
        }

        .search-input:focus {
          outline: none;
          border-color: var(--primary-color);
        }

        .btn-icon {
          background: transparent;
          border: none;
          color: var(--text-secondary);
          cursor: pointer;
          padding: var(--spacing-sm);
          border-radius: var(--radius-md);
          transition: all var(--transition-fast);
        }

        .btn-icon:hover {
          background: var(--surface);
          color: var(--text-primary);
        }

        .notification-btn {
          position: relative;
        }

        .notification-badge {
          position: absolute;
          top: 2px;
          right: 2px;
          background: var(--danger-color);
          color: white;
          font-size: var(--font-size-xs);
          padding: 2px 5px;
          border-radius: var(--radius-full);
          min-width: 16px;
          text-align: center;
        }

        .profile-btn {
          display: flex;
          align-items: center;
          gap: var(--spacing-sm);
          background: transparent;
          border: none;
          cursor: pointer;
          padding: var(--spacing-xs) var(--spacing-sm);
          border-radius: var(--radius-md);
          transition: all var(--transition-fast);
        }

        .profile-btn:hover {
          background: var(--surface);
        }

        .profile-avatar {
          width: 32px;
          height: 32px;
          background: var(--primary-color);
          color: white;
          border-radius: var(--radius-full);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .profile-name {
          font-weight: var(--font-weight-medium);
          color: var(--text-primary);
        }

        .dropdown-menu {
          position: absolute;
          top: calc(100% + var(--spacing-sm));
          right: 0;
          background: var(--background);
          border: 1px solid var(--border-color);
          border-radius: var(--radius-lg);
          box-shadow: var(--shadow-lg);
          min-width: 250px;
          z-index: var(--z-dropdown);
        }

        .notification-menu {
          width: 320px;
        }

        .dropdown-header,
        .dropdown-footer {
          padding: var(--spacing-md);
          border-bottom: 1px solid var(--border-color);
        }

        .dropdown-footer {
          border-bottom: none;
          border-top: 1px solid var(--border-color);
          text-align: center;
        }

        .dropdown-body {
          max-height: 300px;
          overflow-y: auto;
        }

        .notification-item {
          padding: var(--spacing-md);
          border-bottom: 1px solid var(--border-color);
          transition: background var(--transition-fast);
        }

        .notification-item:hover {
          background: var(--surface);
        }

        .notification-item.success {
          border-left: 3px solid var(--success-color);
        }

        .notification-item.warning {
          border-left: 3px solid var(--warning-color);
        }

        .notification-item.info {
          border-left: 3px solid var(--info-color);
        }

        .notification-time {
          font-size: var(--font-size-xs);
          color: var(--text-muted);
        }

        .dropdown-item {
          display: flex;
          align-items: center;
          gap: var(--spacing-sm);
          padding: var(--spacing-sm) var(--spacing-md);
          color: var(--text-primary);
          text-decoration: none;
          transition: background var(--transition-fast);
        }

        .dropdown-item:hover {
          background: var(--surface);
        }

        .sidebar {
          position: fixed;
          top: 64px;
          left: 0;
          bottom: 0;
          background: var(--surface);
          border-right: 1px solid var(--border-color);
          transition: width var(--transition-base);
          z-index: var(--z-sticky);
        }

        .sidebar-open {
          width: 240px;
        }

        .sidebar-closed {
          width: 64px;
        }

        .sidebar-content {
          height: 100%;
          padding: var(--spacing-md);
        }

        .sidebar-nav {
          display: flex;
          flex-direction: column;
          gap: var(--spacing-xs);
        }

        .sidebar-item {
          display: flex;
          align-items: center;
          gap: var(--spacing-md);
          padding: var(--spacing-sm) var(--spacing-md);
          background: transparent;
          border: none;
          border-radius: var(--radius-md);
          color: var(--text-secondary);
          cursor: pointer;
          transition: all var(--transition-fast);
          text-align: left;
          width: 100%;
        }

        .sidebar-item:hover {
          background: var(--background);
          color: var(--text-primary);
        }

        .sidebar-item.active {
          background: var(--primary-color);
          color: white;
        }

        .sidebar-icon {
          font-size: var(--font-size-lg);
          width: 20px;
          text-align: center;
        }

        .sidebar-label {
          flex: 1;
          font-weight: var(--font-weight-medium);
        }

        .sidebar-badge {
          background: var(--danger-color);
          color: white;
          font-size: var(--font-size-xs);
          padding: 2px 6px;
          border-radius: var(--radius-full);
        }

        @media (max-width: 768px) {
          .navbar-toggle {
            display: block;
          }

          .navbar-center {
            display: none;
          }

          .sidebar {
            transform: translateX(-100%);
          }

          .sidebar-open {
            transform: translateX(0);
          }
        }
      `}</style>
    </>
  );
};

export default Navigation;
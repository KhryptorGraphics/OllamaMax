import React from 'react';
import { Nav, Navbar } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faNetworkWired, 
  faTachometerAlt, 
  faServer, 
  faBrain, 
  faExchangeAlt, 
  faSitemap, 
  faChartLine,
  faTimes
} from '@fortawesome/free-solid-svg-icons';

const Sidebar = ({ activeTab, onTabChange, onClose, isOpen = true }) => {
  return (
    <div className={`sidebar ${isOpen ? 'show' : ''}`}>
      <div className="sidebar-content p-3">
        <div className="d-flex justify-content-between align-items-center mb-4">
          <h4 className="mb-0">
            <FontAwesomeIcon icon={faNetworkWired} className="me-2" />
            Ollama Distributed
          </h4>
          <button 
            className="btn btn-sm btn-outline-light d-md-none"
            onClick={onClose}
          >
            <FontAwesomeIcon icon={faTimes} />
          </button>
        </div>
        
        <Nav className="flex-column">
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'dashboard' ? 'active' : ''}
              onClick={() => { onTabChange('dashboard'); onClose(); }}
            >
              <FontAwesomeIcon icon={faTachometerAlt} className="me-2" />
              Dashboard
            </Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'nodes' ? 'active' : ''}
              onClick={() => { onTabChange('nodes'); onClose(); }}
            >
              <FontAwesomeIcon icon={faServer} className="me-2" />
              Nodes
            </Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'models' ? 'active' : ''}
              onClick={() => { onTabChange('models'); onClose(); }}
            >
              <FontAwesomeIcon icon={faBrain} className="me-2" />
              Models
            </Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'transfers' ? 'active' : ''}
              onClick={() => { onTabChange('transfers'); onClose(); }}
            >
              <FontAwesomeIcon icon={faExchangeAlt} className="me-2" />
              Transfers
            </Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'cluster' ? 'active' : ''}
              onClick={() => { onTabChange('cluster'); onClose(); }}
            >
              <FontAwesomeIcon icon={faSitemap} className="me-2" />
              Cluster
            </Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link 
              className={activeTab === 'analytics' ? 'active' : ''}
              onClick={() => { onTabChange('analytics'); onClose(); }}
            >
              <FontAwesomeIcon icon={faChartLine} className="me-2" />
              Analytics
            </Nav.Link>
          </Nav.Item>
        </Nav>
      </div>
    </div>
  );
};

export default Sidebar;
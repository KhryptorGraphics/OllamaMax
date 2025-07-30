import React from 'react';
import { Badge } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faWifi, faWifiSlash, faSync } from '@fortawesome/free-solid-svg-icons';

const WebSocketStatus = ({ isConnected, reconnectAttempts }) => {
  return (
    <div className="websocket-status">
      <Badge 
        bg={isConnected ? 'success' : 'danger'} 
        className="d-flex align-items-center"
      >
        <FontAwesomeIcon 
          icon={isConnected ? faWifi : faWifiSlash} 
          className="me-1" 
        />
        {isConnected ? 'Connected' : `Disconnected ${reconnectAttempts > 0 ? `(${reconnectAttempts}/5)` : ''}`}
        {!isConnected && reconnectAttempts > 0 && (
          <FontAwesomeIcon 
            icon={faSync} 
            className="ms-1 fa-spin" 
          />
        )}
      </Badge>
    </div>
  );
};

export default WebSocketStatus;
import React, { useState, useEffect } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
  faCheckCircle, 
  faExclamationTriangle, 
  faInfoCircle, 
  faTimes,
  faTimesCircle 
} from '@fortawesome/free-solid-svg-icons';

const Alert = ({ 
  type = 'info', 
  title, 
  message, 
  dismissible = true, 
  autoHide = false,
  duration = 5000,
  onDismiss 
}) => {
  const [visible, setVisible] = useState(true);

  const icons = {
    success: faCheckCircle,
    warning: faExclamationTriangle,
    danger: faTimesCircle,
    info: faInfoCircle
  };

  const colors = {
    success: 'success',
    warning: 'warning', 
    danger: 'danger',
    info: 'info'
  };

  useEffect(() => {
    if (autoHide && visible) {
      const timer = setTimeout(() => {
        handleDismiss();
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [autoHide, visible, duration]);

  const handleDismiss = () => {
    setVisible(false);
    if (onDismiss) onDismiss();
  };

  if (!visible) return null;

  return (
    <div className={`alert alert-${colors[type]} d-flex align-items-center alert-dismissible fade show`} role="alert">
      <FontAwesomeIcon icon={icons[type]} className="me-3" size="lg" />
      <div className="flex-grow-1">
        {title && <h6 className="alert-heading mb-1">{title}</h6>}
        {message && <div className="mb-0">{message}</div>}
      </div>
      {dismissible && (
        <button
          type="button"
          className="btn-close"
          aria-label="Close"
          onClick={handleDismiss}
        >
          <FontAwesomeIcon icon={faTimes} />
        </button>
      )}
    </div>
  );
};

export default Alert;
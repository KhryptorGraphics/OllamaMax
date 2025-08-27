import React from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faSpinner } from '@fortawesome/free-solid-svg-icons';

const LoadingSpinner = ({ size = 'md', text = 'Loading...', overlay = false, color = 'primary' }) => {
  const spinnerSize = {
    sm: '1x',
    md: '2x',
    lg: '3x',
    xl: '4x'
  }[size];

  const LoadingComponent = () => (
    <div className={`loading-spinner ${overlay ? 'loading-overlay' : ''}`}>
      <div className="loading-content text-center">
        <FontAwesomeIcon 
          icon={faSpinner} 
          size={spinnerSize}
          className={`fa-spin text-${color} mb-3`}
        />
        {text && <p className="loading-text mb-0">{text}</p>}
      </div>
    </div>
  );

  if (overlay) {
    return (
      <div className="position-fixed top-0 start-0 w-100 h-100 d-flex align-items-center justify-content-center" 
           style={{ zIndex: 9999, backgroundColor: 'rgba(0,0,0,0.5)' }}>
        <LoadingComponent />
      </div>
    );
  }

  return <LoadingComponent />;
};

export default LoadingSpinner;
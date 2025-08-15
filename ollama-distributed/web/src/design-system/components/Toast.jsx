/**
 * Toast Component - OllamaMax Design System
 * 
 * A notification toast component with animations and auto-dismiss functionality.
 */

import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { createPortal } from 'react-dom';
import { tokens } from '../tokens.js';

const Toast = ({
  id,
  type = 'info',
  title,
  message,
  duration = 5000,
  onClose,
  action,
  position = 'top-right'
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isExiting, setIsExiting] = useState(false);

  // Show toast with animation
  useEffect(() => {
    const timer = setTimeout(() => setIsVisible(true), 100);
    return () => clearTimeout(timer);
  }, []);

  // Auto-dismiss timer
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        handleClose();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [duration]);

  // Handle close with animation
  const handleClose = () => {
    setIsExiting(true);
    setTimeout(() => {
      onClose(id);
    }, 300);
  };

  // Toast type configurations
  const typeConfig = {
    success: {
      backgroundColor: tokens.colors.success[500],
      color: tokens.colors.neutral[0],
      icon: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      )
    },
    error: {
      backgroundColor: tokens.colors.error[500],
      color: tokens.colors.neutral[0],
      icon: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      )
    },
    warning: {
      backgroundColor: tokens.colors.warning[500],
      color: tokens.colors.neutral[0],
      icon: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
      )
    },
    info: {
      backgroundColor: tokens.colors.info[500],
      color: tokens.colors.neutral[0],
      icon: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      )
    }
  };

  const config = typeConfig[type];

  // Toast styles
  const toastStyles = {
    display: 'flex',
    alignItems: 'flex-start',
    gap: tokens.spacing[3],
    padding: tokens.spacing[4],
    backgroundColor: config.backgroundColor,
    color: config.color,
    borderRadius: tokens.borderRadius.lg,
    boxShadow: tokens.shadows.lg,
    minWidth: '320px',
    maxWidth: '480px',
    transform: isExiting 
      ? 'translateX(100%)' 
      : isVisible 
        ? 'translateX(0)' 
        : 'translateX(100%)',
    opacity: isExiting ? 0 : isVisible ? 1 : 0,
    transition: `all ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut}`,
    position: 'relative',
    overflow: 'hidden'
  };

  // Progress bar styles (for auto-dismiss)
  const progressBarStyles = {
    position: 'absolute',
    bottom: 0,
    left: 0,
    height: '3px',
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    animation: duration > 0 ? `shrink ${duration}ms linear` : 'none',
    width: '100%'
  };

  // Content styles
  const contentStyles = {
    flex: 1,
    minWidth: 0
  };

  const titleStyles = {
    fontSize: tokens.typography.fontSize.sm[0],
    fontWeight: tokens.typography.fontWeight.semibold,
    marginBottom: title && message ? tokens.spacing[1] : 0,
    lineHeight: tokens.typography.lineHeight.tight
  };

  const messageStyles = {
    fontSize: tokens.typography.fontSize.sm[0],
    lineHeight: tokens.typography.lineHeight.normal,
    opacity: 0.9
  };

  // Close button styles
  const closeButtonStyles = {
    padding: tokens.spacing[1],
    border: 'none',
    backgroundColor: 'transparent',
    color: 'currentColor',
    cursor: 'pointer',
    borderRadius: tokens.borderRadius.sm,
    opacity: 0.7,
    transition: `opacity ${tokens.animation.duration.fast} ${tokens.animation.easing.easeInOut}`,
    ':hover': {
      opacity: 1
    }
  };

  // Action button styles
  const actionButtonStyles = {
    padding: `${tokens.spacing[1]} ${tokens.spacing[2]}`,
    border: '1px solid rgba(255, 255, 255, 0.3)',
    backgroundColor: 'transparent',
    color: 'currentColor',
    borderRadius: tokens.borderRadius.sm,
    fontSize: tokens.typography.fontSize.xs[0],
    fontWeight: tokens.typography.fontWeight.medium,
    cursor: 'pointer',
    marginTop: tokens.spacing[2],
    transition: `all ${tokens.animation.duration.fast} ${tokens.animation.easing.easeInOut}`,
    ':hover': {
      backgroundColor: 'rgba(255, 255, 255, 0.1)'
    }
  };

  return (
    <div style={toastStyles}>
      {/* Icon */}
      <div style={{ flexShrink: 0 }}>
        {config.icon}
      </div>

      {/* Content */}
      <div style={contentStyles}>
        {title && (
          <div style={titleStyles}>
            {title}
          </div>
        )}
        {message && (
          <div style={messageStyles}>
            {message}
          </div>
        )}
        {action && (
          <button
            style={actionButtonStyles}
            onClick={action.onClick}
          >
            {action.label}
          </button>
        )}
      </div>

      {/* Close button */}
      <button
        style={closeButtonStyles}
        onClick={handleClose}
        aria-label="Close notification"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
        </svg>
      </button>

      {/* Progress bar */}
      {duration > 0 && (
        <div style={progressBarStyles} />
      )}
    </div>
  );
};

Toast.propTypes = {
  id: PropTypes.string.isRequired,
  type: PropTypes.oneOf(['success', 'error', 'warning', 'info']),
  title: PropTypes.string,
  message: PropTypes.string.isRequired,
  duration: PropTypes.number,
  onClose: PropTypes.func.isRequired,
  action: PropTypes.shape({
    label: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired
  }),
  position: PropTypes.oneOf(['top-right', 'top-left', 'bottom-right', 'bottom-left'])
};

// Toast Container Component
export const ToastContainer = ({ toasts, position = 'top-right' }) => {
  // Position styles
  const positionStyles = {
    'top-right': {
      top: tokens.spacing[4],
      right: tokens.spacing[4]
    },
    'top-left': {
      top: tokens.spacing[4],
      left: tokens.spacing[4]
    },
    'bottom-right': {
      bottom: tokens.spacing[4],
      right: tokens.spacing[4]
    },
    'bottom-left': {
      bottom: tokens.spacing[4],
      left: tokens.spacing[4]
    }
  };

  const containerStyles = {
    position: 'fixed',
    zIndex: tokens.zIndex.toast,
    display: 'flex',
    flexDirection: 'column',
    gap: tokens.spacing[2],
    pointerEvents: 'none',
    ...positionStyles[position]
  };

  const toastWrapperStyles = {
    pointerEvents: 'auto'
  };

  if (!toasts || toasts.length === 0) return null;

  const containerContent = (
    <div style={containerStyles}>
      {toasts.map((toast) => (
        <div key={toast.id} style={toastWrapperStyles}>
          <Toast {...toast} />
        </div>
      ))}
    </div>
  );

  return createPortal(containerContent, document.body);
};

ToastContainer.propTypes = {
  toasts: PropTypes.arrayOf(PropTypes.object),
  position: PropTypes.oneOf(['top-right', 'top-left', 'bottom-right', 'bottom-left'])
};

// CSS animations
const styles = `
  @keyframes shrink {
    from { width: 100%; }
    to { width: 0%; }
  }
`;

// Inject styles
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.textContent = styles;
  document.head.appendChild(styleSheet);
}

export default Toast;

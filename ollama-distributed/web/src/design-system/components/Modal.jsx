/**
 * Modal Component - OllamaMax Design System
 * 
 * A flexible modal component with accessibility features and animations.
 */

import React, { useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import { createPortal } from 'react-dom';
import { tokens } from '../tokens.js';
import Button from './Button.jsx';

const Modal = ({
  isOpen,
  onClose,
  title,
  children,
  size = 'md',
  closeOnOverlayClick = true,
  closeOnEscape = true,
  showCloseButton = true,
  footer,
  className = '',
  ...props
}) => {
  const modalRef = useRef(null);
  const previousActiveElement = useRef(null);

  // Handle escape key
  useEffect(() => {
    const handleEscape = (event) => {
      if (closeOnEscape && event.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, closeOnEscape, onClose]);

  // Handle focus management
  useEffect(() => {
    if (isOpen) {
      // Store the currently focused element
      previousActiveElement.current = document.activeElement;
      
      // Focus the modal
      if (modalRef.current) {
        modalRef.current.focus();
      }
      
      // Prevent body scroll
      document.body.style.overflow = 'hidden';
    } else {
      // Restore focus to the previously focused element
      if (previousActiveElement.current) {
        previousActiveElement.current.focus();
      }
      
      // Restore body scroll
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  // Handle overlay click
  const handleOverlayClick = (event) => {
    if (closeOnOverlayClick && event.target === event.currentTarget) {
      onClose();
    }
  };

  // Handle focus trap
  const handleKeyDown = (event) => {
    if (event.key === 'Tab') {
      const focusableElements = modalRef.current?.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      
      if (focusableElements && focusableElements.length > 0) {
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];
        
        if (event.shiftKey) {
          if (document.activeElement === firstElement) {
            event.preventDefault();
            lastElement.focus();
          }
        } else {
          if (document.activeElement === lastElement) {
            event.preventDefault();
            firstElement.focus();
          }
        }
      }
    }
  };

  if (!isOpen) return null;

  // Overlay styles
  const overlayStyles = {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: tokens.zIndex.modal,
    padding: tokens.spacing[4],
    animation: `fadeIn ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut}`
  };

  // Size variants
  const sizeStyles = {
    sm: {
      maxWidth: '400px',
      width: '100%'
    },
    md: {
      maxWidth: '500px',
      width: '100%'
    },
    lg: {
      maxWidth: '700px',
      width: '100%'
    },
    xl: {
      maxWidth: '900px',
      width: '100%'
    },
    full: {
      maxWidth: '95vw',
      maxHeight: '95vh',
      width: '100%',
      height: '100%'
    }
  };

  // Modal styles
  const modalStyles = {
    backgroundColor: tokens.colors.neutral[0],
    borderRadius: tokens.borderRadius.lg,
    boxShadow: tokens.shadows['2xl'],
    display: 'flex',
    flexDirection: 'column',
    maxHeight: '90vh',
    outline: 'none',
    animation: `slideUp ${tokens.animation.duration.normal} ${tokens.animation.easing.easeOut}`,
    ...sizeStyles[size]
  };

  // Header styles
  const headerStyles = {
    padding: `${tokens.spacing[6]} ${tokens.spacing[6]} ${tokens.spacing[4]}`,
    borderBottom: `1px solid ${tokens.colors.neutral[200]}`,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between'
  };

  const titleStyles = {
    fontSize: tokens.typography.fontSize.lg[0],
    fontWeight: tokens.typography.fontWeight.semibold,
    color: tokens.colors.neutral[900],
    margin: 0
  };

  // Body styles
  const bodyStyles = {
    padding: tokens.spacing[6],
    flex: 1,
    overflow: 'auto'
  };

  // Footer styles
  const footerStyles = {
    padding: `${tokens.spacing[4]} ${tokens.spacing[6]} ${tokens.spacing[6]}`,
    borderTop: `1px solid ${tokens.colors.neutral[200]}`,
    display: 'flex',
    justifyContent: 'flex-end',
    gap: tokens.spacing[3]
  };

  // Close button styles
  const closeButtonStyles = {
    padding: tokens.spacing[2],
    border: 'none',
    backgroundColor: 'transparent',
    color: tokens.colors.neutral[400],
    cursor: 'pointer',
    borderRadius: tokens.borderRadius.md,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: `color ${tokens.animation.duration.fast} ${tokens.animation.easing.easeInOut}`,
    ':hover': {
      color: tokens.colors.neutral[600]
    }
  };

  const modalContent = (
    <div style={overlayStyles} onClick={handleOverlayClick}>
      <div
        ref={modalRef}
        style={modalStyles}
        className={className}
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        {...props}
      >
        {/* Header */}
        {(title || showCloseButton) && (
          <div style={headerStyles}>
            {title && (
              <h2 id="modal-title" style={titleStyles}>
                {title}
              </h2>
            )}
            {showCloseButton && (
              <button
                style={closeButtonStyles}
                onClick={onClose}
                aria-label="Close modal"
              >
                <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                </svg>
              </button>
            )}
          </div>
        )}

        {/* Body */}
        <div style={bodyStyles}>
          {children}
        </div>

        {/* Footer */}
        {footer && (
          <div style={footerStyles}>
            {footer}
          </div>
        )}
      </div>
    </div>
  );

  // Render modal in portal
  return createPortal(modalContent, document.body);
};

Modal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  title: PropTypes.string,
  children: PropTypes.node.isRequired,
  size: PropTypes.oneOf(['sm', 'md', 'lg', 'xl', 'full']),
  closeOnOverlayClick: PropTypes.bool,
  closeOnEscape: PropTypes.bool,
  showCloseButton: PropTypes.bool,
  footer: PropTypes.node,
  className: PropTypes.string
};

// Confirmation Modal Component
export const ConfirmModal = ({
  isOpen,
  onClose,
  onConfirm,
  title = 'Confirm Action',
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'primary',
  loading = false
}) => {
  const handleConfirm = async () => {
    await onConfirm();
  };

  const footer = (
    <>
      <Button variant="secondary" onClick={onClose} disabled={loading}>
        {cancelText}
      </Button>
      <Button 
        variant={variant} 
        onClick={handleConfirm} 
        loading={loading}
        disabled={loading}
      >
        {confirmText}
      </Button>
    </>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      size="sm"
      footer={footer}
      closeOnOverlayClick={!loading}
      closeOnEscape={!loading}
    >
      <p style={{
        color: tokens.colors.neutral[700],
        lineHeight: tokens.typography.lineHeight.relaxed,
        margin: 0
      }}>
        {message}
      </p>
    </Modal>
  );
};

ConfirmModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired,
  title: PropTypes.string,
  message: PropTypes.string.isRequired,
  confirmText: PropTypes.string,
  cancelText: PropTypes.string,
  variant: PropTypes.oneOf(['primary', 'danger', 'warning']),
  loading: PropTypes.bool
};

// CSS animations
const styles = `
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  
  @keyframes slideUp {
    from { 
      opacity: 0;
      transform: translateY(20px) scale(0.95);
    }
    to { 
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
`;

// Inject styles
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.textContent = styles;
  document.head.appendChild(styleSheet);
}

export default Modal;
